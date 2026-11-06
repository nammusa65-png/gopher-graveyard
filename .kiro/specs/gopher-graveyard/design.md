# Design — Gopher Graveyard

## Components

- `internal/gopher` — client core, written from scratch:
  `ParseURL` (gopher://host[:port]/<type><selector>), `Fetch`
  (TCP dial, send selector + optional TAB + search, read to EOF),
  `ParseMenu` (tab-separated menu lines → typed items with URLs).
- `internal/gopherserver` — server core: per-connection goroutine,
  reads one selector line, resolves it under a content root
  (path-cleaned, traversal-safe), serves `gophermap` files verbatim,
  documents as-is, and generated listings for bare directories.
  Menus/text are terminated with a lone "." line.
- `cmd/gopherd` — the server binary (`-addr`, `-root`, `-host`).
- `cmd/graveyard` — the web gateway binary. Serves the static UI and:
  - `GET /api/fetch?url=<gopher-url>[&search=<terms>]` → JSON
    `{kind: menu|text, items, text, raw}` or 502 `{error}` (rendered
    by the UI as a tombstone).
  - `GET /api/check?url=<gopher-url>` → `{alive: bool}` for the
    graveyard map, with a short timeout.
- `cmd/graveyard/web` — dependency-free HTML/CSS/JS UI. Graveyard
  theme, address bar + history, per-type icons, search prompt for
  type-7 items, séance mode (raw view), tombstone error view.

## Data flow

Browser → HTTP → gateway (Go) → raw TCP :70 → any Gopher server.
The gateway is required because browsers cannot open raw TCP
sockets nor gopher:// URLs.

## Safety

- Response size cap (8 MiB), dial deadline, per-request context
  timeout on every fetch.
- `gopherd` path resolution is confined to its content root.
- The gateway is a demo bridge: it fetches only the Gopher protocol
  on the host/port named in the URL the user typed.
