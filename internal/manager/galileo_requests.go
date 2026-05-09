package manager

// Galileo fork: media request worker bootstrap.

import (
	"context"

	"github.com/stashapp/stash/internal/manager/config"
	"github.com/stashapp/stash/pkg/logger"
	"github.com/stashapp/stash/pkg/requests"
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

	store := s.Database.MediaRequest
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

type galileoWorkerLogger struct{}

func (galileoWorkerLogger) Infof(format string, args ...any)  { logger.Infof(format, args...) }
func (galileoWorkerLogger) Errorf(format string, args ...any) { logger.Errorf(format, args...) }
