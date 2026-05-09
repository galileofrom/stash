package requests

import (
	"errors"
	"strings"
	"time"
)

// Config is the runtime configuration for the requests subsystem,
// resolved from the stash config file or environment.
// ProtocolPreference controls how releases are ordered when both Usenet and
// torrent results come back from Prowlarr. "usenet" or "torrent" hard-prefer
// one protocol over the other; "auto" leaves it to RankWeights so each release
// competes on its own score.
type ProtocolPreference string

const (
	PreferUsenet  ProtocolPreference = "usenet"
	PreferTorrent ProtocolPreference = "torrent"
	PreferAuto    ProtocolPreference = "auto"
)

// RankWeights are the per-attribute multipliers used when ranking release
// candidates. Higher values pull the corresponding attribute toward the top
// of the list. Negative values are valid and demote a release.
type RankWeights struct {
	UsenetBoost   float64 // additive score for usenet releases (default 50)
	TorrentBoost  float64 // additive score for torrent releases (default 0)
	SeederWeight  float64 // multiplier on seeder count (default 1)
	FreeleechBoost float64 // additive score when indexer flags freeleech (default 5)
	AgeDecayDays  float64 // releases older than this lose score (default 30)
}

func DefaultRankWeights() RankWeights {
	return RankWeights{
		UsenetBoost:    50,
		TorrentBoost:   0,
		SeederWeight:   1,
		FreeleechBoost: 5,
		AgeDecayDays:   30,
	}
}

type Config struct {
	ProwlarrEnabled    bool
	ProwlarrURL        string
	ProwlarrAPIKey     string
	ProwlarrCategories []int
	ProwlarrIndexerIDs []int
	ProwlarrTimeout    time.Duration
	LibraryPath        string
	RequireApproval    bool

	// Ranking + filtering
	PreferProtocol ProtocolPreference
	MinSeeders     int
	MaxSizeBytes   int64
	RankWeights    RankWeights
}

func (c Config) Validate() error {
	if !c.ProwlarrEnabled {
		return nil
	}
	if strings.TrimSpace(c.ProwlarrURL) == "" {
		return errors.New("prowlarr.url is required when prowlarr.enabled is true")
	}
	if strings.TrimSpace(c.ProwlarrAPIKey) == "" {
		return errors.New("prowlarr.api_key is required when prowlarr.enabled is true")
	}
	if strings.TrimSpace(c.LibraryPath) == "" {
		return errors.New("requests.library_path is required: completed downloads land here for stash to scan")
	}
	return nil
}

func (c Config) ProwlarrClientConfig() ProwlarrConfig {
	return ProwlarrConfig{
		BaseURL:    c.ProwlarrURL,
		APIKey:     c.ProwlarrAPIKey,
		Categories: c.ProwlarrCategories,
		IndexerIDs: c.ProwlarrIndexerIDs,
		Timeout:    c.ProwlarrTimeout,
	}
}
