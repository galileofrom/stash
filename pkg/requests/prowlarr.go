package requests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ProwlarrConfig holds connection details for a Prowlarr instance.
type ProwlarrConfig struct {
	BaseURL string
	APIKey  string
	// Categories defaults to {6000} (XXX) when empty.
	Categories []int
	// IndexerIDs limits search to specific indexers when set.
	IndexerIDs []int
	Timeout    time.Duration
}

// ProwlarrClient is a thin wrapper around the Prowlarr v1 REST API.
type ProwlarrClient struct {
	cfg  ProwlarrConfig
	http *http.Client
}

func NewProwlarrClient(cfg ProwlarrConfig) *ProwlarrClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if len(cfg.Categories) == 0 {
		cfg.Categories = []int{6000}
	}
	return &ProwlarrClient{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// ProwlarrRelease mirrors the relevant subset of Prowlarr's search response.
type ProwlarrRelease struct {
	GUID         string    `json:"guid"`
	Title        string    `json:"title"`
	IndexerID    int       `json:"indexerId"`
	Indexer      string    `json:"indexer"`
	Size         int64     `json:"size"`
	Seeders      int       `json:"seeders"`
	Leechers     int       `json:"leechers"`
	PublishDate  time.Time `json:"publishDate"`
	DownloadURL  string    `json:"downloadUrl"`
	MagnetURL    string    `json:"magnetUrl"`
	InfoURL      string    `json:"infoUrl"`
	Protocol     string    `json:"protocol"`
	Categories   []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"categories"`
}

// Search runs a Prowlarr release search for the given query string.
func (c *ProwlarrClient) Search(ctx context.Context, query string) ([]ProwlarrRelease, error) {
	q := url.Values{}
	q.Set("query", query)
	q.Set("type", "search")
	for _, cat := range c.cfg.Categories {
		q.Add("categories", strconv.Itoa(cat))
	}
	for _, id := range c.cfg.IndexerIDs {
		q.Add("indexerIds", strconv.Itoa(id))
	}

	endpoint := strings.TrimRight(c.cfg.BaseURL, "/") + "/api/v1/search?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", c.cfg.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("prowlarr search %d: %s", resp.StatusCode, string(body))
	}

	var releases []ProwlarrRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("decode prowlarr response: %w", err)
	}
	return releases, nil
}

// Grab forwards a release to the configured download client via Prowlarr.
// indexerID and guid identify the release; Prowlarr looks up the rest.
func (c *ProwlarrClient) Grab(ctx context.Context, indexerID int, guid string) error {
	payload := map[string]any{
		"guid":      guid,
		"indexerId": indexerID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	endpoint := strings.TrimRight(c.cfg.BaseURL, "/") + "/api/v1/search"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("prowlarr grab %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// Ping verifies connectivity and credentials.
func (c *ProwlarrClient) Ping(ctx context.Context) error {
	endpoint := strings.TrimRight(c.cfg.BaseURL, "/") + "/api/v1/system/status"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", c.cfg.APIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("prowlarr ping %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
