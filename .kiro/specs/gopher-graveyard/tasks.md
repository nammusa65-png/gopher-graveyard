# Tasks — Gopher Graveyard

## Phase 1 — core resurrection (scaffold)
- [x] Gopher client core: URL parse, TCP fetch, menu parse + tests
- [x] Gopher server core: gophermap serving, generated listings,
      traversal guard + tests
- [x] Web gateway: /api/fetch, /api/check, static UI
- [x] Graveyard map with alive/dead check; tombstone error view
- [x] Séance mode (raw protocol view)
- [x] Sample gopherhole content (about / why Gopher died / docs)

## Phase 2 — to be built with Kiro (spec-driven sessions)
- [ ] Last-seen timestamps and latency on the graveyard map
      (persist check results server-side, show "last answered …")
- [ ] Bookmarks ("plots"): localStorage-backed saved holes
- [ ] Image items (type I) rendered inline via a safe proxy route
- [ ] Keyboard navigation (j/k/Enter) for menu items
- [ ] English Devpost write-up of how Kiro was used (specs, steering,
      agent hooks) — written from the actual Phase 2 session history
- [ ] Deployment recipe (single container: gateway + gopherd)

## Phase 3 — submission
- [ ] Demo video (< 3 min) following the brief's script outline
- [ ] Public repo polish: screenshots, LICENSE present, /.kiro in place
- [ ] Submit on Devpost before 2026-12-04 14:00 PT
