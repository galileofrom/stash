package requests

import (
	"testing"
	"time"
)

func release(protocol string, seeders int, size int64, age time.Duration) *Release {
	return &Release{
		Protocol:    protocol,
		Seeders:     seeders,
		Size:        size,
		PublishDate: time.Now().Add(-age),
	}
}

func TestRanker_PreferUsenet_PutsUsenetFirst(t *testing.T) {
	cfg := Config{PreferProtocol: PreferUsenet, RankWeights: DefaultRankWeights()}
	r := NewRanker(cfg)

	in := []*Release{
		release("torrent", 1000, 1<<30, 0),
		release("usenet", 0, 1<<30, 0),
	}

	out := r.Rank(in)
	if len(out) != 2 {
		t.Fatalf("expected 2 releases, got %d", len(out))
	}
	if !isUsenet(out[0].Protocol) {
		t.Fatalf("usenet should rank first, got %q", out[0].Protocol)
	}
}

func TestRanker_PreferTorrent_PutsTorrentFirst(t *testing.T) {
	cfg := Config{PreferProtocol: PreferTorrent, RankWeights: DefaultRankWeights()}
	r := NewRanker(cfg)

	in := []*Release{
		release("usenet", 0, 1<<30, 0),
		release("torrent", 5, 1<<30, 0),
	}

	out := r.Rank(in)
	if !isTorrent(out[0].Protocol) {
		t.Fatalf("torrent should rank first, got %q", out[0].Protocol)
	}
}

func TestRanker_MinSeeders_DropsLowSeederTorrents(t *testing.T) {
	cfg := Config{MinSeeders: 5, RankWeights: DefaultRankWeights()}
	r := NewRanker(cfg)

	in := []*Release{
		release("torrent", 2, 1<<30, 0),
		release("torrent", 50, 1<<30, 0),
	}
	out := r.Rank(in)
	if len(out) != 1 {
		t.Fatalf("expected 1 release after filter, got %d", len(out))
	}
	if out[0].Seeders != 50 {
		t.Fatalf("unexpected seeder count after filter: %d", out[0].Seeders)
	}
}

func TestRanker_MinSeeders_IgnoresUsenet(t *testing.T) {
	cfg := Config{MinSeeders: 5, RankWeights: DefaultRankWeights()}
	r := NewRanker(cfg)

	in := []*Release{release("usenet", 0, 1<<30, 0)}
	out := r.Rank(in)
	if len(out) != 1 {
		t.Fatalf("min seeders must not apply to usenet, got %d", len(out))
	}
}

func TestRanker_MaxSizeBytes_DropsOversize(t *testing.T) {
	cfg := Config{MaxSizeBytes: 5 << 30, RankWeights: DefaultRankWeights()}
	r := NewRanker(cfg)

	in := []*Release{
		release("torrent", 100, 10<<30, 0),
		release("torrent", 100, 1<<30, 0),
	}
	out := r.Rank(in)
	if len(out) != 1 {
		t.Fatalf("expected 1 release after size filter, got %d", len(out))
	}
}

func TestRanker_AgeDecay_DemotesOldRelease(t *testing.T) {
	cfg := Config{
		PreferProtocol: PreferAuto,
		RankWeights:    RankWeights{UsenetBoost: 50, AgeDecayDays: 30, SeederWeight: 1},
	}
	r := NewRanker(cfg)

	fresh := release("usenet", 0, 1<<30, 0)
	stale := release("usenet", 0, 1<<30, 200*24*time.Hour)
	out := r.Rank([]*Release{stale, fresh})
	if out[0] != fresh {
		t.Fatalf("fresh release should outrank stale one")
	}
}
