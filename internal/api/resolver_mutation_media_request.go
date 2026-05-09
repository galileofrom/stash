package api

import (
	"context"
	"errors"
	"strconv"

	"github.com/stashapp/stash/internal/manager"
	"github.com/stashapp/stash/pkg/requests"
	"github.com/stashapp/stash/pkg/sqlite"
)

// Galileo fork: mutation resolvers for the media request workflow.

func (r *mutationResolver) MediaRequestCreate(ctx context.Context, input MediaRequestCreateInput) (*MediaRequest, error) {
	store := mediaRequestStore()

	req := &requests.MediaRequest{
		Title:  input.Title,
		Status: requests.StatusPending,
	}
	if input.Studio != nil {
		req.Studio = *input.Studio
	}
	if input.ExternalID != nil {
		req.ExternalID = *input.ExternalID
	}
	if input.ExternalSource != nil {
		req.ExternalSource = requests.ExternalSource(*input.ExternalSource)
	}
	if input.Notes != nil {
		req.Notes = *input.Notes
	}

	if err := r.withTxn(ctx, func(ctx context.Context) error {
		_, err := store.CreateRequest(ctx, req)
		return err
	}); err != nil {
		return nil, err
	}
	return mediaRequestToGQL(req, nil), nil
}

func (r *mutationResolver) MediaRequestSearch(ctx context.Context, id string) (*MediaRequest, error) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	svc, err := newRequestService()
	if err != nil {
		return nil, err
	}

	var (
		req      *requests.MediaRequest
		releases []*requests.Release
	)
	if err := r.withTxn(ctx, func(ctx context.Context) error {
		releases, err = svc.SearchReleases(ctx, idInt)
		if err != nil {
			return err
		}
		req, err = mediaRequestStore().GetRequest(ctx, idInt)
		return err
	}); err != nil {
		return nil, err
	}
	return mediaRequestToGQL(req, releases), nil
}

func (r *mutationResolver) MediaRequestApprove(ctx context.Context, releaseID string) (*MediaRequest, error) {
	idInt, err := strconv.Atoi(releaseID)
	if err != nil {
		return nil, err
	}

	svc, err := newRequestService()
	if err != nil {
		return nil, err
	}

	store := mediaRequestStore()
	var (
		req      *requests.MediaRequest
		releases []*requests.Release
	)
	if err := r.withTxn(ctx, func(ctx context.Context) error {
		if err := svc.Approve(ctx, idInt); err != nil {
			return err
		}
		rel, err := store.GetRelease(ctx, idInt)
		if err != nil {
			return err
		}
		req, err = store.GetRequest(ctx, rel.RequestID)
		if err != nil {
			return err
		}
		releases, err = store.ListReleases(ctx, rel.RequestID)
		return err
	}); err != nil {
		return nil, err
	}
	return mediaRequestToGQL(req, releases), nil
}

func (r *mutationResolver) MediaRequestReject(ctx context.Context, id string, reason *string) (*MediaRequest, error) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	store := mediaRequestStore()
	var req *requests.MediaRequest
	if err := r.withTxn(ctx, func(ctx context.Context) error {
		req, err = store.GetRequest(ctx, idInt)
		if err != nil {
			return err
		}
		req.Status = requests.StatusRejected
		if reason != nil {
			req.Notes = *reason
		}
		return store.UpdateRequest(ctx, req)
	}); err != nil {
		return nil, err
	}
	return mediaRequestToGQL(req, nil), nil
}

func (r *mutationResolver) MediaRequestDestroy(ctx context.Context, id string) (bool, error) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return false, err
	}

	store := mediaRequestStore()
	if err := r.withTxn(ctx, func(ctx context.Context) error {
		return store.DeleteRequest(ctx, idInt)
	}); err != nil {
		return false, err
	}
	return true, nil
}

// newRequestService builds a Service from current stash config + database.
// Returns an error if Prowlarr is disabled or required config is missing.
func newRequestService() (*requests.Service, error) {
	cfg, err := loadRequestsConfig()
	if err != nil {
		return nil, err
	}
	if !cfg.ProwlarrEnabled {
		return nil, errors.New("prowlarr is disabled in config (set prowlarr.enabled=true)")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return requests.NewService(cfg, mediaRequestStore()), nil
}

func mediaRequestStore() *sqlite.MediaRequestStore {
	return manager.GetInstance().Database.MediaRequest
}
