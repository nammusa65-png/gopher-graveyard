# 🪦 Gopher Graveyard

**Browsers killed Gopher — and then dropped support for it entirely.
Gopher Graveyard resurrects Gopherspace inside a modern browser.**

Enter any `gopher://` URL and browse live Gopher servers through a
haunted, polished web UI. Dead hosts become tombstones. A from-scratch
Go Gopher server in the same repo serves our own gopherhole, so both
ends of the 1991 protocol (RFC 1436) are implemented here.

> Kiroween 2026 entry — category: **Resurrection**.

## Why this exists

Gopher (1991, University of Minnesota) organised the early internet
as text menus. The Web won, and browsers quietly removed `gopher://`
support — today a normal browser *cannot* open Gopherspace at all.
But it is not empty: a January 2026 Veronica-2 census counted **411
active servers and ~5.86 million selectors**, kept alive by
enthusiasts (Floodgap since 1999, SDF, and hundreds of hobby holes).

The protocol is beautifully small: connect to TCP port 70, send one
selector line, read the reply, hang up. No state. This project
re-implements it from scratch and gives it a modern face.

## Run it

```sh
# Web gateway + UI (http://localhost:8070)
go run ./cmd/graveyard -addr :8070

# Our own Gopher server (serves ./content on port 7070 for local use;
# use :70 with sufficient privileges for the real thing)
go run ./cmd/gopherd -addr :7070 -root content -host localhost
```

Then open http://localhost:8070 and dig: start at
`gopher://gopher.floodgap.com`, search Gopherspace with Veronica-2,
or visit the University of Minnesota's original server and pay your
respects at its tombstone.

## Layout

- `internal/gopher` — Gopher client core (URL parse, TCP fetch, menu parse)
- `internal/gopherserver` — Gopher server core (gophermaps, listings)
- `cmd/graveyard` — web gateway (`/api/fetch`, `/api/check`) + UI
- `cmd/gopherd` — standalone Gopher server
- `content/` — the project's own gopherhole
- `.kiro/specs/gopher-graveyard/` — Kiro specs (requirements / design / tasks)

## Development & how Kiro is used

Development is spec-driven with Kiro: the specs under `.kiro/specs/`
define the requirements, design, and task list, and Phase 2 features
(see `tasks.md`) are implemented in Kiro sessions against those
specs. The initial Phase 1 scaffold was prototyped with an AI
assistant before the Kiro sessions; the Devpost write-up will
describe the actual Kiro workflow (specs, steering, and the features
built with it) from session history.

## Tests

```sh
go test ./...
go vet ./...
```

## License

MIT — see [LICENSE](LICENSE).
