package requests

import (
	"errors"
	"strings"
	"time"
)

// Config is the runtime configuration for the requests subsystem,
// resolved from the stash config file or environment.
type Config struct {
	ProwlarrEnabled        bool
	ProwlarrURL            string
	ProwlarrAPIKey         string
	ProwlarrCategories     []int
	ProwlarrIndexerIDs     []int
	ProwlarrTimeout        time.Duration
	LibraryPath            string
	RequireApproval        bool
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
