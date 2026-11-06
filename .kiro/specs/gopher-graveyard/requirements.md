# Requirements — Gopher Graveyard

Kiroween 2026 entry, category: **Resurrection**.
A modern web gateway that resurrects the Gopher protocol (RFC 1436):
browsers dropped `gopher://`, so this app speaks Gopher over raw TCP
and renders Gopherspace in a polished, haunted web UI.

## User stories

1. As a visitor, I can enter any `gopher://` URL and browse the live
   Gopher menu it returns, rendered as a modern page.
2. As a visitor, I can open text documents (type 0), follow menus
   (type 1), and run searches (type 7, e.g. Veronica-2).
3. As a visitor, when a host does not answer, I see a tombstone with
   the failure reason — dead hosts are part of the exhibit.
4. As a visitor, I can open the "graveyard map" of known holes and
   see at a glance which are still alive.
5. As a visitor, I can toggle "séance mode" to see the raw protocol
   lines next to the rendered UI, proving the traffic is real Gopher.
6. As an operator, I can run `gopherd` to serve my own gopherhole
   (gophermaps + text files + generated listings) from scratch —
   both ends of the protocol live in this repository.

## Acceptance criteria

- Fetching uses raw TCP (port 70 by default), sends exactly one
  selector line terminated by CRLF, and reads until close (RFC 1436).
- Menu parsing handles item types 0, 1, 3, 7, 9, i, h and stops at
  the terminating "." line.
- The gateway caps response size (8 MiB) and applies timeouts so a
  hostile or hung server cannot wedge it.
- `gopherd` refuses path traversal outside its content root.
- All core packages have unit tests (`go test ./...`).

## Non-goals (v1)

- Accounts, bookmark sync, posting to Gopher.
- Gopher+ support; binary content beyond download links.
- Crawling or indexing all of Gopherspace.
