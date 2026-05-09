package requests

import (
	"sort"
	"strings"
	"time"
)

// Ranker scores and filters Prowlarr releases according to user preferences.
type Ranker struct {
	cfg Config
}

func NewRanker(cfg Config) Ranker {
	if cfg.RankWeights == (RankWeights{}) {
		cfg.RankWeights = DefaultRankWeights()
	}
	if cfg.PreferProtocol == "" {
		cfg.PreferProtocol = PreferAuto
	}
	return Ranker{cfg: cfg}
}

// Rank filters out releases that violate hard limits (min seeders, max size)
// and returns the remainder sorted by descending score.
func (r Ranker) Rank(releases []*Release) []*Release {
	out := make([]*Release, 0, len(releases))
	for _, rel := range releases {
		if !r.passesFilters(rel) {
			continue
		}
		out = append(out, rel)
	}

	sort.SliceStable(out, func(i, j int) bool {
		return r.score(out[i]) > r.score(out[j])
	})
	return out
}

func (r Ranker) passesFilters(rel *Release) bool {
	if isTorrent(rel.Protocol) && r.cfg.MinSeeders > 0 && rel.Seeders < r.cfg.MinSeeders {
		return false
	}
	if r.cfg.MaxSizeBytes > 0 && rel.Size > r.cfg.MaxSizeBytes {
		return false
	}
	return true
}

func (r Ranker) score(rel *Release) float64 {
	w := r.cfg.RankWeights
	score := 0.0

	switch {
	case isUsenet(rel.Protocol):
		score += w.UsenetBoost
	case isTorrent(rel.Protocol):
		score += w.TorrentBoost + float64(rel.Seeders)*w.SeederWeight
	}

	// Hard preference overrides additive boost. Pin preferred protocol to a
	// large positive offset so any non-preferred candidate sits below.
	switch r.cfg.PreferProtocol {
	case PreferUsenet:
		if isUsenet(rel.Protocol) {
			score += 10000
		}
	case PreferTorrent:
		if isTorrent(rel.Protocol) {
			score += 10000
		}
	}

	if !rel.PublishDate.IsZero() && w.AgeDecayDays > 0 {
		ageDays := time.Since(rel.PublishDate).Hours() / 24
		if ageDays > w.AgeDecayDays {
			score -= ageDays - w.AgeDecayDays
		}
	}

	return score
}

func isUsenet(protocol string) bool {
	return strings.EqualFold(protocol, "usenet")
}

func isTorrent(protocol string) bool {
	return strings.EqualFold(protocol, "torrent")
}
