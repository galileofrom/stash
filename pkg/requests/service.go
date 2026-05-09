package requests

import (
	"context"
	"errors"
)

// ErrNotImplemented marks methods that exist as part of the public surface
// but whose persistence/wiring lands in a follow-up commit.
var ErrNotImplemented = errors.New("requests: not implemented yet")

// Repository abstracts persistence so the service stays testable without
// a real sqlite database.
type Repository interface {
	CreateRequest(ctx context.Context, r *MediaRequest) (int, error)
	UpdateRequest(ctx context.Context, r *MediaRequest) error
	GetRequest(ctx context.Context, id int) (*MediaRequest, error)
	ListRequests(ctx context.Context, status Status) ([]*MediaRequest, error)

	UpsertReleases(ctx context.Context, requestID int, releases []*Release) error
	ListReleases(ctx context.Context, requestID int) ([]*Release, error)
	GetRelease(ctx context.Context, id int) (*Release, error)

	CreateDownload(ctx context.Context, d *Download) (int, error)
	UpdateDownload(ctx context.Context, d *Download) error
	ListDownloads(ctx context.Context, status DownloadStatus) ([]*Download, error)
}

// Service is the high-level entry point for the requests workflow.
type Service struct {
	cfg      Config
	repo     Repository
	prowlarr *ProwlarrClient
	ranker   Ranker
}

func NewService(cfg Config, repo Repository) *Service {
	var pw *ProwlarrClient
	if cfg.ProwlarrEnabled {
		pw = NewProwlarrClient(cfg.ProwlarrClientConfig())
	}
	return &Service{cfg: cfg, repo: repo, prowlarr: pw, ranker: NewRanker(cfg)}
}

// SearchReleases queries Prowlarr for a candidate request and persists
// the returned releases against it.
func (s *Service) SearchReleases(ctx context.Context, requestID int) ([]*Release, error) {
	if s.prowlarr == nil {
		return nil, errors.New("prowlarr not configured")
	}
	req, err := s.repo.GetRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}

	query := req.Title
	if req.Studio != "" {
		query = req.Studio + " " + req.Title
	}

	raw, err := s.prowlarr.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	releases := make([]*Release, 0, len(raw))
	for _, r := range raw {
		category := ""
		if len(r.Categories) > 0 {
			category = r.Categories[0].Name
		}
		releases = append(releases, &Release{
			RequestID:   requestID,
			Indexer:     r.Indexer,
			IndexerID:   r.IndexerID,
			GUID:        r.GUID,
			Title:       r.Title,
			DownloadURL: r.DownloadURL,
			MagnetURL:   r.MagnetURL,
			InfoURL:     r.InfoURL,
			Size:        r.Size,
			Seeders:     r.Seeders,
			Leechers:    r.Leechers,
			PublishDate: r.PublishDate,
			Category:    category,
			Protocol:    r.Protocol,
		})
	}

	releases = s.ranker.Rank(releases)

	if err := s.repo.UpsertReleases(ctx, requestID, releases); err != nil {
		return nil, err
	}
	return releases, nil
}

// Approve marks a release for download and forwards it to Prowlarr's grab endpoint.
func (s *Service) Approve(ctx context.Context, releaseID int) error {
	if s.prowlarr == nil {
		return errors.New("prowlarr not configured")
	}
	rel, err := s.repo.GetRelease(ctx, releaseID)
	if err != nil {
		return err
	}
	if err := s.prowlarr.Grab(ctx, rel.IndexerID, rel.GUID); err != nil {
		return err
	}

	req, err := s.repo.GetRequest(ctx, rel.RequestID)
	if err != nil {
		return err
	}
	req.Status = StatusDownloading
	return s.repo.UpdateRequest(ctx, req)
}
