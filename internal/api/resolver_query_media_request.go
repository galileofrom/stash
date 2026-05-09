package api

import (
	"context"
	"strconv"

	"github.com/stashapp/stash/pkg/requests"
)

// Galileo fork: query resolvers for the media request workflow.

func (r *queryResolver) FindMediaRequest(ctx context.Context, id string) (*MediaRequest, error) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	store := mediaRequestStore()
	var req *requests.MediaRequest
	var releases []*requests.Release
	if err := r.withReadTxn(ctx, func(ctx context.Context) error {
		req, err = store.GetRequest(ctx, idInt)
		if err != nil {
			return err
		}
		releases, err = store.ListReleases(ctx, idInt)
		return err
	}); err != nil {
		return nil, err
	}
	return mediaRequestToGQL(req, releases), nil
}

func (r *queryResolver) FindMediaRequests(ctx context.Context, filter *MediaRequestFilter) ([]*MediaRequest, error) {
	var status requests.Status
	if filter != nil && filter.Status != nil {
		status = mediaRequestStatusFromGQL(*filter.Status)
	}

	store := mediaRequestStore()
	var reqs []*requests.MediaRequest
	if err := r.withReadTxn(ctx, func(ctx context.Context) error {
		var err error
		reqs, err = store.ListRequests(ctx, status)
		return err
	}); err != nil {
		return nil, err
	}

	out := make([]*MediaRequest, len(reqs))
	for i, req := range reqs {
		// Skip releases here to keep list lightweight; a per-request fetch
		// hydrates them on detail view via FindMediaRequest.
		out[i] = mediaRequestToGQL(req, nil)
	}
	return out, nil
}
