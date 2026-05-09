package requests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

// QBittorrentConfig connects galileo to a qBittorrent Web UI.
type QBittorrentConfig struct {
	BaseURL  string
	Username string
	Password string
	Timeout  time.Duration
}

// QBittorrentClient is a minimal qBittorrent v2 Web API client. Only the
// endpoints galileo needs are implemented (login + torrent listing/info).
type QBittorrentClient struct {
	cfg    QBittorrentConfig
	http   *http.Client
	mu     sync.Mutex
	signedIn bool
}

func NewQBittorrentClient(cfg QBittorrentConfig) (*QBittorrentClient, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &QBittorrentClient{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout, Jar: jar},
	}, nil
}

// Login authenticates against qBittorrent and stores the session cookie.
// Calling more than once is safe; the qBit session cookie has a long TTL.
func (c *QBittorrentClient) Login(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.signedIn {
		return nil
	}

	form := url.Values{}
	form.Set("username", c.cfg.Username)
	form.Set("password", c.cfg.Password)

	endpoint := strings.TrimRight(c.cfg.BaseURL, "/") + "/api/v2/auth/login"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.cfg.BaseURL)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK || strings.TrimSpace(string(body)) != "Ok." {
		return fmt.Errorf("qbittorrent login failed: %d %s", resp.StatusCode, string(body))
	}
	c.signedIn = true
	return nil
}

// QBitTorrent mirrors the relevant subset of /api/v2/torrents/info output.
type QBitTorrent struct {
	Hash       string  `json:"hash"`
	Name       string  `json:"name"`
	State      string  `json:"state"`
	Progress   float64 `json:"progress"`
	ETA        int     `json:"eta"`
	ContentPath string `json:"content_path"`
	SavePath   string  `json:"save_path"`
	Size       int64   `json:"size"`
	Completed  int64   `json:"completed"`
}

// TorrentByHash returns the torrent matching hash, or nil if not present.
func (c *QBittorrentClient) TorrentByHash(ctx context.Context, hash string) (*QBitTorrent, error) {
	if err := c.Login(ctx); err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("hashes", strings.ToLower(hash))
	endpoint := strings.TrimRight(c.cfg.BaseURL, "/") + "/api/v2/torrents/info?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("qbittorrent torrents/info %d: %s", resp.StatusCode, string(body))
	}

	var torrents []QBitTorrent
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		return nil, err
	}
	if len(torrents) == 0 {
		return nil, nil
	}
	return &torrents[0], nil
}

// IsComplete reports whether qBit considers this state finished or seeding.
func (t QBitTorrent) IsComplete() bool {
	switch strings.ToLower(t.State) {
	case "uploading", "stalledup", "queuedup", "pauseduP", "pausedup", "forcedup", "checkingup":
		return true
	}
	return t.Progress >= 1.0
}
