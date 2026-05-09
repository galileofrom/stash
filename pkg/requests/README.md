# requests

Galileo fork: media request workflow.

## Goal

Let a user search external metadata, find downloadable releases via
Prowlarr, approve specific releases, and have completed downloads land
in a stash-watched library path for normal stash scanning to pick up.

## Layout

- `models.go` — domain types (`MediaRequest`, `Release`, `Download`)
- `config.go` — runtime config + validation
- `prowlarr.go` — Prowlarr v1 REST client (`/api/v1/search`, grab, ping)
- `ranker.go` — score + filter releases by protocol preference, seeders, size, age
- `service.go` — orchestration (search → rank → persist → approve → grab)

## Ranking

`PreferProtocol` controls protocol preference: `usenet`, `torrent`, or
`auto` (additive `RankWeights` only). `MinSeeders` and `MaxSizeBytes`
hard-filter torrents; usenet is exempt from min seeders. Age decay
demotes releases older than `RankWeights.AgeDecayDays`.

## Status

Foundation only. No persistence implementation, no GraphQL resolvers,
no UI, no background worker yet. Tracking in feature/media-requests.

## Schema

See `pkg/sqlite/migrations/90000_media_requests.up.sql`. Migration
numbers in this fork start at 90000 to leave room for upstream
migrations < 90000 without rename conflicts on merge.

## Open questions

- Download client: route through Prowlarr's grab (simpler) vs talk to
  qBit/SAB directly (more state, better progress tracking).
- Metadata source priority: StashDB vs ThePornDB vs both. Reuse stash's
  existing scraper machinery rather than reimplementing.
- Approval model: per-request approval, per-release approval, or both.
