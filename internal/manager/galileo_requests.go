package manager

// Galileo fork: media request worker bootstrap.

import (
	"context"

	"github.com/stashapp/stash/internal/manager/config"
	"github.com/stashapp/stash/pkg/logger"
	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/requests"
	"github.com/stashapp/stash/pkg/txn"
)

// startMediaRequestsWorker boots the background worker if Prowlarr is enabled.
// It runs for the lifetime of ctx; postInit is called once per Manager,
// so a single goroutine per process is correct.
func (s *Manager) startMediaRequestsWorker(ctx context.Context) {
	if !s.Config.GetProwlarrEnabled() {
		logger.Debugf("media requests: prowlarr disabled, worker not started")
		return
	}

	cfg := buildRequestsConfig(s.Config)
	if err := cfg.Validate(); err != nil {
		logger.Warnf("media requests: configuration invalid, worker not started: %v", err)
		return
	}

	// Wrap the store so worker calls automatically run inside a stash txn,
	// matching the behaviour of every other repository access in the codebase.
	store := newTxnRequestStore(s.Database.MediaRequest, s.Repository.TxnManager)
	prowlarr := requests.NewProwlarrClient(cfg.ProwlarrClientConfig())

	// qBittorrent client is optional; absence simply means downloads cannot
	// be polled and will sit in the queued state.
	var qbit *requests.QBittorrentClient
	if qbCfg, ok := buildQBittorrentConfig(s.Config); ok {
		client, err := requests.NewQBittorrentClient(qbCfg)
		if err != nil {
			logger.Warnf("media requests: qbittorrent client init failed: %v", err)
		} else {
			qbit = client
		}
	}

	worker := requests.NewWorker(requests.WorkerConfig{
		Repo:        store,
		Prowlarr:    prowlarr,
		QBittorrent: qbit,
		Importer:    nil,
		Scanner:     nil, // wired in a follow-up commit
		LibraryPath: cfg.LibraryPath,
		Logger:      &galileoWorkerLogger{},
	})

	go worker.Run(ctx)
	logger.Infof("media requests: worker started (prowlarr=%s, library=%s)", cfg.ProwlarrURL, cfg.LibraryPath)
}

// buildRequestsConfig mirrors the loader in internal/api/media_request_helpers.go,
// duplicated here to avoid an import cycle (api -> manager -> api).
func buildRequestsConfig(c *config.Config) requests.Config {
	cfg := requests.Config{
		ProwlarrEnabled:    c.GetProwlarrEnabled(),
		ProwlarrURL:        c.GetProwlarrURL(),
		ProwlarrAPIKey:     c.GetProwlarrAPIKey(),
		ProwlarrCategories: c.GetProwlarrCategories(),
		ProwlarrIndexerIDs: c.GetProwlarrIndexerIDs(),
		LibraryPath:        c.GetRequestsLibraryPath(),
		RequireApproval:    c.GetRequestsRequireApproval(),
		PreferProtocol:     requests.ProtocolPreference(c.GetRequestsPreferProtocol()),
		MinSeeders:         c.GetRequestsMinSeeders(),
		MaxSizeBytes:       c.GetRequestsMaxSizeBytes(),
	}
	cfg.RankWeights = requests.DefaultRankWeights()
	return cfg
}

// buildQBittorrentConfig is best-effort; returns false if connection details
// are missing so callers can decide to skip the client.
func buildQBittorrentConfig(c *config.Config) (requests.QBittorrentConfig, bool) {
	url := c.GetRequestsQBittorrentURL()
	if url == "" {
		return requests.QBittorrentConfig{}, false
	}
	return requests.QBittorrentConfig{
		BaseURL:  url,
		Username: c.GetRequestsQBittorrentUser(),
		Password: c.GetRequestsQBittorrentPass(),
	}, true
}

// txnRequestStore wraps a sqlite.MediaRequestStore so its methods run inside
// a transaction. The worker calls store methods without an outer ctx already
// in a txn (it owns its own goroutine), so we open one per call.
type txnRequestStore struct {
	inner requests.Repository
	txnMgr models.TxnManager
}

func newTxnRequestStore(inner requests.Repository, txnMgr models.TxnManager) *txnRequestStore {
	return &txnRequestStore{inner: inner, txnMgr: txnMgr}
}

func (s *txnRequestStore) inRead(ctx context.Context, fn func(ctx context.Context) error) error {
	return txn.WithReadTxn(ctx, s.txnMgr, fn)
}

func (s *txnRequestStore) inWrite(ctx context.Context, fn func(ctx context.Context) error) error {
	return txn.WithTxn(ctx, s.txnMgr, fn)
}

func (s *txnRequestStore) CreateRequest(ctx context.Context, r *requests.MediaRequest) (id int, err error) {
	err = s.inWrite(ctx, func(ctx context.Context) error {
		id, err = s.inner.CreateRequest(ctx, r)
		return err
	})
	return
}

func (s *txnRequestStore) UpdateRequest(ctx context.Context, r *requests.MediaRequest) error {
	return s.inWrite(ctx, func(ctx context.Context) error {
		return s.inner.UpdateRequest(ctx, r)
	})
}

func (s *txnRequestStore) GetRequest(ctx context.Context, id int) (out *requests.MediaRequest, err error) {
	err = s.inRead(ctx, func(ctx context.Context) error {
		out, err = s.inner.GetRequest(ctx, id)
		return err
	})
	return
}

func (s *txnRequestStore) ListRequests(ctx context.Context, status requests.Status) (out []*requests.MediaRequest, err error) {
	err = s.inRead(ctx, func(ctx context.Context) error {
		out, err = s.inner.ListRequests(ctx, status)
		return err
	})
	return
}

func (s *txnRequestStore) UpsertReleases(ctx context.Context, requestID int, releases []*requests.Release) error {
	return s.inWrite(ctx, func(ctx context.Context) error {
		return s.inner.UpsertReleases(ctx, requestID, releases)
	})
}

func (s *txnRequestStore) ListReleases(ctx context.Context, requestID int) (out []*requests.Release, err error) {
	err = s.inRead(ctx, func(ctx context.Context) error {
		out, err = s.inner.ListReleases(ctx, requestID)
		return err
	})
	return
}

func (s *txnRequestStore) GetRelease(ctx context.Context, id int) (out *requests.Release, err error) {
	err = s.inRead(ctx, func(ctx context.Context) error {
		out, err = s.inner.GetRelease(ctx, id)
		return err
	})
	return
}

func (s *txnRequestStore) CreateDownload(ctx context.Context, d *requests.Download) (id int, err error) {
	err = s.inWrite(ctx, func(ctx context.Context) error {
		id, err = s.inner.CreateDownload(ctx, d)
		return err
	})
	return
}

func (s *txnRequestStore) UpdateDownload(ctx context.Context, d *requests.Download) error {
	return s.inWrite(ctx, func(ctx context.Context) error {
		return s.inner.UpdateDownload(ctx, d)
	})
}

func (s *txnRequestStore) ListDownloads(ctx context.Context, status requests.DownloadStatus) (out []*requests.Download, err error) {
	err = s.inRead(ctx, func(ctx context.Context) error {
		out, err = s.inner.ListDownloads(ctx, status)
		return err
	})
	return
}

type galileoWorkerLogger struct{}

func (galileoWorkerLogger) Infof(format string, args ...any)  { logger.Infof(format, args...) }
func (galileoWorkerLogger) Errorf(format string, args ...any) { logger.Errorf(format, args...) }
