package sqlite

// Galileo fork: persistence layer for the media request workflow.
// Implements pkg/requests.Repository on top of the stash sqlite database
// using the same tableMgr/repository plumbing the rest of the project uses.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"

	"github.com/stashapp/stash/pkg/requests"
)

const (
	mediaRequestTable          = "media_requests"
	mediaRequestReleaseTable   = "media_request_releases"
	mediaRequestDownloadsTable = "media_request_downloads"
)

type mediaRequestRow struct {
	ID             int            `db:"id" goqu:"skipinsert"`
	Title          string         `db:"title"`
	Studio         sql.NullString `db:"studio"`
	ExternalID     sql.NullString `db:"external_id"`
	ExternalSource sql.NullString `db:"external_source"`
	MetadataJSON   sql.NullString `db:"metadata_json"`
	Status         string         `db:"status"`
	RequestedBy    sql.NullString `db:"requested_by"`
	Notes          sql.NullString `db:"notes"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

func (r *mediaRequestRow) fromDomain(m *requests.MediaRequest) {
	r.ID = m.ID
	r.Title = m.Title
	r.Studio = nullString(m.Studio)
	r.ExternalID = nullString(m.ExternalID)
	r.ExternalSource = nullString(string(m.ExternalSource))
	r.MetadataJSON = nullString(m.MetadataJSON)
	r.Status = string(m.Status)
	r.RequestedBy = nullString(m.RequestedBy)
	r.Notes = nullString(m.Notes)
	r.CreatedAt = m.CreatedAt
	r.UpdatedAt = m.UpdatedAt
}

func (r *mediaRequestRow) toDomain() *requests.MediaRequest {
	return &requests.MediaRequest{
		ID:             r.ID,
		Title:          r.Title,
		Studio:         r.Studio.String,
		ExternalID:     r.ExternalID.String,
		ExternalSource: requests.ExternalSource(r.ExternalSource.String),
		MetadataJSON:   r.MetadataJSON.String,
		Status:         requests.Status(r.Status),
		RequestedBy:    r.RequestedBy.String,
		Notes:          r.Notes.String,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

type mediaReleaseRow struct {
	ID          int            `db:"id" goqu:"skipinsert"`
	RequestID   int            `db:"request_id"`
	Indexer     string         `db:"indexer"`
	IndexerID   sql.NullInt64  `db:"indexer_id"`
	GUID        string         `db:"guid"`
	Title       string         `db:"title"`
	DownloadURL sql.NullString `db:"download_url"`
	MagnetURL   sql.NullString `db:"magnet_url"`
	InfoURL     sql.NullString `db:"info_url"`
	Size        sql.NullInt64  `db:"size"`
	Seeders     sql.NullInt64  `db:"seeders"`
	Leechers    sql.NullInt64  `db:"leechers"`
	PublishDate sql.NullTime   `db:"publish_date"`
	Category    sql.NullString `db:"category"`
	Protocol    sql.NullString `db:"protocol"`
	RawJSON     sql.NullString `db:"raw_json"`
	Selected    bool           `db:"selected"`
	GrabStatus  sql.NullString `db:"grab_status"`
	GrabError   sql.NullString `db:"grab_error"`
	CreatedAt   time.Time      `db:"created_at"`
}

func (r *mediaReleaseRow) fromDomain(m *requests.Release) {
	r.ID = m.ID
	r.RequestID = m.RequestID
	r.Indexer = m.Indexer
	if m.IndexerID > 0 {
		r.IndexerID = sql.NullInt64{Int64: int64(m.IndexerID), Valid: true}
	}
	r.GUID = m.GUID
	r.Title = m.Title
	r.DownloadURL = nullString(m.DownloadURL)
	r.MagnetURL = nullString(m.MagnetURL)
	r.InfoURL = nullString(m.InfoURL)
	r.Size = sql.NullInt64{Int64: m.Size, Valid: m.Size > 0}
	r.Seeders = sql.NullInt64{Int64: int64(m.Seeders), Valid: true}
	r.Leechers = sql.NullInt64{Int64: int64(m.Leechers), Valid: true}
	if !m.PublishDate.IsZero() {
		r.PublishDate = sql.NullTime{Time: m.PublishDate, Valid: true}
	}
	r.Category = nullString(m.Category)
	r.Protocol = nullString(m.Protocol)
	r.Selected = m.Selected
	r.GrabStatus = nullString(m.GrabStatus)
	r.GrabError = nullString(m.GrabError)
	r.CreatedAt = m.CreatedAt
}

func (r *mediaReleaseRow) toDomain() *requests.Release {
	out := &requests.Release{
		ID:          r.ID,
		RequestID:   r.RequestID,
		Indexer:     r.Indexer,
		IndexerID:   int(r.IndexerID.Int64),
		GUID:        r.GUID,
		Title:       r.Title,
		DownloadURL: r.DownloadURL.String,
		MagnetURL:   r.MagnetURL.String,
		InfoURL:     r.InfoURL.String,
		Size:        r.Size.Int64,
		Seeders:     int(r.Seeders.Int64),
		Leechers:    int(r.Leechers.Int64),
		Category:    r.Category.String,
		Protocol:    r.Protocol.String,
		Selected:    r.Selected,
		GrabStatus:  r.GrabStatus.String,
		GrabError:   r.GrabError.String,
		CreatedAt:   r.CreatedAt,
	}
	if r.PublishDate.Valid {
		out.PublishDate = r.PublishDate.Time
	}
	return out
}

type mediaDownloadRow struct {
	ID             int            `db:"id" goqu:"skipinsert"`
	RequestID      int            `db:"request_id"`
	ReleaseID      int            `db:"release_id"`
	DownloadClient sql.NullString `db:"download_client"`
	DownloadID     sql.NullString `db:"download_id"`
	Status         string         `db:"status"`
	Progress       float64        `db:"progress"`
	ETASeconds     sql.NullInt64  `db:"eta_seconds"`
	OutputPath     sql.NullString `db:"output_path"`
	Error          sql.NullString `db:"error"`
	StartedAt      sql.NullTime   `db:"started_at"`
	CompletedAt    sql.NullTime   `db:"completed_at"`
	ImportedAt     sql.NullTime   `db:"imported_at"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

// MediaRequestStore implements requests.Repository.
type MediaRequestStore struct{}

func NewMediaRequestStore() *MediaRequestStore {
	return &MediaRequestStore{}
}

// compile-time check: this store satisfies the requests.Repository interface.
var _ requests.Repository = (*MediaRequestStore)(nil)

func (s *MediaRequestStore) CreateRequest(ctx context.Context, req *requests.MediaRequest) (int, error) {
	if req.CreatedAt.IsZero() {
		req.CreatedAt = time.Now()
	}
	req.UpdatedAt = time.Now()
	if req.Status == "" {
		req.Status = requests.StatusPending
	}

	var row mediaRequestRow
	row.fromDomain(req)
	id, err := mediaRequestTableMgr.insertID(ctx, row)
	if err != nil {
		return 0, err
	}
	req.ID = id
	return id, nil
}

func (s *MediaRequestStore) UpdateRequest(ctx context.Context, req *requests.MediaRequest) error {
	req.UpdatedAt = time.Now()
	var row mediaRequestRow
	row.fromDomain(req)
	return mediaRequestTableMgr.updateByID(ctx, req.ID, row)
}

func (s *MediaRequestStore) GetRequest(ctx context.Context, id int) (*requests.MediaRequest, error) {
	q := dialect.From(mediaRequestTable).Where(goqu.C("id").Eq(id)).Limit(1)
	var row mediaRequestRow
	if err := s.queryRow(ctx, q, &row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("media request %d not found", id)
		}
		return nil, err
	}
	return row.toDomain(), nil
}

func (s *MediaRequestStore) ListRequests(ctx context.Context, status requests.Status) ([]*requests.MediaRequest, error) {
	q := dialect.From(mediaRequestTable).Order(goqu.C("created_at").Desc())
	if status != "" {
		q = q.Where(goqu.C("status").Eq(string(status)))
	}

	var rows []mediaRequestRow
	if err := s.queryRows(ctx, q, &rows); err != nil {
		return nil, err
	}

	out := make([]*requests.MediaRequest, len(rows))
	for i := range rows {
		out[i] = rows[i].toDomain()
	}
	return out, nil
}

// DeleteRequest removes a request and (via FK ON DELETE CASCADE) its releases
// and downloads.
func (s *MediaRequestStore) DeleteRequest(ctx context.Context, id int) error {
	q := dialect.Delete(mediaRequestTable).Where(goqu.C("id").Eq(id))
	return s.run(ctx, q)
}

func (s *MediaRequestStore) UpsertReleases(ctx context.Context, requestID int, releases []*requests.Release) error {
	// Strategy: replace-all per request. Search results from Prowlarr are
	// regenerated on every search, so wiping and reinserting is simpler than
	// reconciling by GUID and avoids stale rows.
	delQ := dialect.Delete(mediaRequestReleaseTable).Where(goqu.C("request_id").Eq(requestID))
	if err := s.run(ctx, delQ); err != nil {
		return fmt.Errorf("clearing releases for request %d: %w", requestID, err)
	}

	for _, rel := range releases {
		rel.RequestID = requestID
		if rel.CreatedAt.IsZero() {
			rel.CreatedAt = time.Now()
		}
		var row mediaReleaseRow
		row.fromDomain(rel)
		id, err := mediaReleaseTableMgr.insertID(ctx, row)
		if err != nil {
			return fmt.Errorf("inserting release %q: %w", rel.GUID, err)
		}
		rel.ID = id
	}
	return nil
}

func (s *MediaRequestStore) ListReleases(ctx context.Context, requestID int) ([]*requests.Release, error) {
	q := dialect.From(mediaRequestReleaseTable).Where(goqu.C("request_id").Eq(requestID))

	var rows []mediaReleaseRow
	if err := s.queryRows(ctx, q, &rows); err != nil {
		return nil, err
	}
	out := make([]*requests.Release, len(rows))
	for i := range rows {
		out[i] = rows[i].toDomain()
	}
	return out, nil
}

func (s *MediaRequestStore) GetRelease(ctx context.Context, id int) (*requests.Release, error) {
	q := dialect.From(mediaRequestReleaseTable).Where(goqu.C("id").Eq(id)).Limit(1)
	var row mediaReleaseRow
	if err := s.queryRow(ctx, q, &row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("media release %d not found", id)
		}
		return nil, err
	}
	return row.toDomain(), nil
}

func (s *MediaRequestStore) CreateDownload(ctx context.Context, d *requests.Download) (int, error) {
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}
	d.UpdatedAt = time.Now()
	if d.Status == "" {
		d.Status = requests.DownloadQueued
	}

	row := downloadFromDomain(d)
	id, err := mediaDownloadTableMgr.insertID(ctx, row)
	if err != nil {
		return 0, err
	}
	d.ID = id
	return id, nil
}

func (s *MediaRequestStore) UpdateDownload(ctx context.Context, d *requests.Download) error {
	d.UpdatedAt = time.Now()
	row := downloadFromDomain(d)
	return mediaDownloadTableMgr.updateByID(ctx, d.ID, row)
}

func (s *MediaRequestStore) ListDownloads(ctx context.Context, status requests.DownloadStatus) ([]*requests.Download, error) {
	q := dialect.From(mediaRequestDownloadsTable)
	if status != "" {
		q = q.Where(goqu.C("status").Eq(string(status)))
	}

	var rows []mediaDownloadRow
	if err := s.queryRows(ctx, q, &rows); err != nil {
		return nil, err
	}
	out := make([]*requests.Download, len(rows))
	for i := range rows {
		out[i] = downloadToDomain(&rows[i])
	}
	return out, nil
}

// queryRow runs a goqu select that returns at most one row and decodes it
// into dst. dst must be a pointer to a struct.
func (s *MediaRequestStore) queryRow(ctx context.Context, q *goqu.SelectDataset, dst interface{}) error {
	stmt, args, err := q.Prepared(true).ToSQL()
	if err != nil {
		return err
	}
	tx, err := getTx(ctx)
	if err != nil {
		return err
	}
	return tx.GetContext(ctx, dst, stmt, args...)
}

func (s *MediaRequestStore) queryRows(ctx context.Context, q *goqu.SelectDataset, dst interface{}) error {
	stmt, args, err := q.Prepared(true).ToSQL()
	if err != nil {
		return err
	}
	tx, err := getTx(ctx)
	if err != nil {
		return err
	}
	return tx.SelectContext(ctx, dst, stmt, args...)
}

// run executes a non-query (delete/update) goqu statement against the active txn.
func (s *MediaRequestStore) run(ctx context.Context, q sqler) error {
	_, err := exec(ctx, q)
	return err
}

// nullString turns "" into a SQL NULL.
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func downloadFromDomain(d *requests.Download) mediaDownloadRow {
	row := mediaDownloadRow{
		ID:             d.ID,
		RequestID:      d.RequestID,
		ReleaseID:      d.ReleaseID,
		DownloadClient: nullString(d.DownloadClient),
		DownloadID:     nullString(d.DownloadID),
		Status:         string(d.Status),
		Progress:       d.Progress,
		OutputPath:     nullString(d.OutputPath),
		Error:          nullString(d.Error),
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
	if d.ETASeconds > 0 {
		row.ETASeconds = sql.NullInt64{Int64: int64(d.ETASeconds), Valid: true}
	}
	if d.StartedAt != nil {
		row.StartedAt = sql.NullTime{Time: *d.StartedAt, Valid: true}
	}
	if d.CompletedAt != nil {
		row.CompletedAt = sql.NullTime{Time: *d.CompletedAt, Valid: true}
	}
	if d.ImportedAt != nil {
		row.ImportedAt = sql.NullTime{Time: *d.ImportedAt, Valid: true}
	}
	return row
}

func downloadToDomain(r *mediaDownloadRow) *requests.Download {
	d := &requests.Download{
		ID:             r.ID,
		RequestID:      r.RequestID,
		ReleaseID:      r.ReleaseID,
		DownloadClient: r.DownloadClient.String,
		DownloadID:     r.DownloadID.String,
		Status:         requests.DownloadStatus(r.Status),
		Progress:       r.Progress,
		ETASeconds:     int(r.ETASeconds.Int64),
		OutputPath:     r.OutputPath.String,
		Error:          r.Error.String,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
	if r.StartedAt.Valid {
		t := r.StartedAt.Time
		d.StartedAt = &t
	}
	if r.CompletedAt.Valid {
		t := r.CompletedAt.Time
		d.CompletedAt = &t
	}
	if r.ImportedAt.Valid {
		t := r.ImportedAt.Time
		d.ImportedAt = &t
	}
	return d
}
