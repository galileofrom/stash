package api

import (
	"strconv"
	"time"

	"github.com/stashapp/stash/internal/manager/config"
	"github.com/stashapp/stash/pkg/requests"
)

// mediaRequestToGQL converts a domain MediaRequest plus its releases into the
// generated GraphQL model. releases may be nil to skip release hydration.
func mediaRequestToGQL(req *requests.MediaRequest, releases []*requests.Release) *MediaRequest {
	if req == nil {
		return nil
	}
	out := &MediaRequest{
		ID:        strconv.Itoa(req.ID),
		Title:     req.Title,
		Status:    mediaRequestStatusToGQL(req.Status),
		CreatedAt: req.CreatedAt,
		UpdatedAt: req.UpdatedAt,
	}
	if req.Studio != "" {
		s := req.Studio
		out.Studio = &s
	}
	if req.ExternalID != "" {
		s := req.ExternalID
		out.ExternalID = &s
	}
	if req.ExternalSource != "" {
		s := string(req.ExternalSource)
		out.ExternalSource = &s
	}
	if req.RequestedBy != "" {
		s := req.RequestedBy
		out.RequestedBy = &s
	}
	if req.Notes != "" {
		s := req.Notes
		out.Notes = &s
	}
	out.Releases = make([]*MediaRelease, len(releases))
	for i, rel := range releases {
		out.Releases[i] = mediaReleaseToGQL(rel)
	}
	return out
}

func mediaReleaseToGQL(rel *requests.Release) *MediaRelease {
	if rel == nil {
		return nil
	}
	out := &MediaRelease{
		ID:        strconv.Itoa(rel.ID),
		RequestID: strconv.Itoa(rel.RequestID),
		Indexer:   rel.Indexer,
		GUID:      rel.GUID,
		Title:     rel.Title,
		Size:      rel.Size,
		Seeders:   rel.Seeders,
		Leechers:  rel.Leechers,
		Protocol:  protocolToGQL(rel.Protocol),
		Selected:  rel.Selected,
		CreatedAt: rel.CreatedAt,
	}
	if rel.IndexerID > 0 {
		v := rel.IndexerID
		out.IndexerID = &v
	}
	if rel.DownloadURL != "" {
		v := rel.DownloadURL
		out.DownloadURL = &v
	}
	if rel.MagnetURL != "" {
		v := rel.MagnetURL
		out.MagnetURL = &v
	}
	if rel.InfoURL != "" {
		v := rel.InfoURL
		out.InfoURL = &v
	}
	if !rel.PublishDate.IsZero() {
		t := rel.PublishDate
		out.PublishDate = &t
	}
	if rel.Category != "" {
		v := rel.Category
		out.Category = &v
	}
	if rel.GrabStatus != "" {
		v := rel.GrabStatus
		out.GrabStatus = &v
	}
	if rel.GrabError != "" {
		v := rel.GrabError
		out.GrabError = &v
	}
	return out
}

func mediaRequestStatusToGQL(s requests.Status) MediaRequestStatus {
	switch s {
	case requests.StatusApproved:
		return MediaRequestStatusApproved
	case requests.StatusDownloading:
		return MediaRequestStatusDownloading
	case requests.StatusImported:
		return MediaRequestStatusImported
	case requests.StatusRejected:
		return MediaRequestStatusRejected
	case requests.StatusFailed:
		return MediaRequestStatusFailed
	default:
		return MediaRequestStatusPending
	}
}

func mediaRequestStatusFromGQL(s MediaRequestStatus) requests.Status {
	switch s {
	case MediaRequestStatusApproved:
		return requests.StatusApproved
	case MediaRequestStatusDownloading:
		return requests.StatusDownloading
	case MediaRequestStatusImported:
		return requests.StatusImported
	case MediaRequestStatusRejected:
		return requests.StatusRejected
	case MediaRequestStatusFailed:
		return requests.StatusFailed
	default:
		return requests.StatusPending
	}
}

func protocolToGQL(p string) MediaReleaseProtocol {
	switch p {
	case "usenet":
		return MediaReleaseProtocolUsenet
	case "torrent":
		return MediaReleaseProtocolTorrent
	default:
		return MediaReleaseProtocolUnknown
	}
}

// loadRequestsConfig reads the requests/Prowlarr keys from stash config and
// builds a runtime requests.Config.
func loadRequestsConfig() (requests.Config, error) {
	c := config.GetInstance()

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

	if t := c.GetProwlarrTimeoutSeconds(); t > 0 {
		cfg.ProwlarrTimeout = time.Duration(t) * time.Second
	}

	cfg.RankWeights = requests.DefaultRankWeights()
	if v := c.GetRequestsRankUsenetBoost(); v != 0 {
		cfg.RankWeights.UsenetBoost = v
	}
	if v := c.GetRequestsRankTorrentBoost(); v != 0 {
		cfg.RankWeights.TorrentBoost = v
	}
	if v := c.GetRequestsRankSeederWeight(); v != 0 {
		cfg.RankWeights.SeederWeight = v
	}
	if v := c.GetRequestsRankFreeleechBoost(); v != 0 {
		cfg.RankWeights.FreeleechBoost = v
	}
	if v := c.GetRequestsRankAgeDecayDays(); v != 0 {
		cfg.RankWeights.AgeDecayDays = v
	}
	return cfg, nil
}
