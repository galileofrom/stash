// Package requests implements the media request workflow for the Galileo
// fork of stash. It lets users search external metadata sources, find
// downloadable releases via Prowlarr, approve/reject them, and track
// downloads through to library import.
package requests

import "time"

type Status string

const (
	StatusPending     Status = "pending"     // searched, awaiting approval
	StatusApproved    Status = "approved"    // user approved a release
	StatusDownloading Status = "downloading" // sent to download client
	StatusImported    Status = "imported"    // landed in stash library
	StatusRejected    Status = "rejected"    // user rejected
	StatusFailed      Status = "failed"      // grab or import failure
)

type ExternalSource string

const (
	SourceStashBox ExternalSource = "stash-box"
	SourceTPDB     ExternalSource = "tpdb"
	SourceManual   ExternalSource = "manual"
)

type MediaRequest struct {
	ID             int            `json:"id"`
	Title          string         `json:"title"`
	Studio         string         `json:"studio,omitempty"`
	ExternalID     string         `json:"external_id,omitempty"`
	ExternalSource ExternalSource `json:"external_source,omitempty"`
	MetadataJSON   string         `json:"metadata_json,omitempty"`
	Status         Status         `json:"status"`
	RequestedBy    string         `json:"requested_by,omitempty"`
	Notes          string         `json:"notes,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type Release struct {
	ID           int       `json:"id"`
	RequestID    int       `json:"request_id"`
	Indexer      string    `json:"indexer"`
	IndexerID    int       `json:"indexer_id,omitempty"`
	GUID         string    `json:"guid"`
	Title        string    `json:"title"`
	DownloadURL  string    `json:"download_url,omitempty"`
	MagnetURL    string    `json:"magnet_url,omitempty"`
	InfoURL      string    `json:"info_url,omitempty"`
	Size         int64     `json:"size"`
	Seeders      int       `json:"seeders"`
	Leechers     int       `json:"leechers"`
	PublishDate  time.Time `json:"publish_date"`
	Category     string    `json:"category,omitempty"`
	Protocol     string    `json:"protocol,omitempty"`
	Selected     bool      `json:"selected"`
	GrabStatus   string    `json:"grab_status,omitempty"`
	GrabError    string    `json:"grab_error,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type DownloadStatus string

const (
	DownloadQueued      DownloadStatus = "queued"
	DownloadDownloading DownloadStatus = "downloading"
	DownloadCompleted   DownloadStatus = "completed"
	DownloadImported    DownloadStatus = "imported"
	DownloadFailed      DownloadStatus = "failed"
)

type Download struct {
	ID             int            `json:"id"`
	RequestID      int            `json:"request_id"`
	ReleaseID      int            `json:"release_id"`
	DownloadClient string         `json:"download_client,omitempty"`
	DownloadID     string         `json:"download_id,omitempty"`
	Status         DownloadStatus `json:"status"`
	Progress       float64        `json:"progress"`
	ETASeconds     int            `json:"eta_seconds,omitempty"`
	OutputPath     string         `json:"output_path,omitempty"`
	Error          string         `json:"error,omitempty"`
	StartedAt      *time.Time     `json:"started_at,omitempty"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
	ImportedAt     *time.Time     `json:"imported_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}
