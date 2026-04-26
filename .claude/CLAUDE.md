# HTML Comments

- For elements with begin/end tags (`<div>`, `<svg>`, etc.):
  place describing comments inside, between the tags.
- For self-closing elements (`<img>`, `<input>`, etc.):
  place describing comments before the element.

# Pixel Values

- Use even numbers for all pixel values in CSS and SVG
  (points, positions, dimensions, stroke widths),
  so scaling down by 0.5 produces no half-pixel artifacts.

# Dev Server

- Run `go run cmd/main-yugioh-card-editor/main.go` to start the server at http://localhost:20808

# SQLite Access

- `sqlite3` CLI is not available on this machine.
- To query `data/yugioh.db`, modify and run `cmd/query-sqlite-cardid/query_sqlite_cardid.go`.
  That file can be freely extended with extra queries and reused between sessions.
- `data/yugioh.db` can be safely deleted and re-created by running
  `go run cmd/crawl-konami-db-full/crawl_konami_db_full.go`.
  Konami card pages are cached under `data/html_cache/`, so re-crawls do not re-hit Konami.

# Testing

- Playwright tests live in `web_tests/` (including `package.json`,
  `package-lock.json`, `playwright.config.ts`, and `node_modules/`).
  Run tests with `cd web_tests && npx playwright test`.

# konami.CardID

- `konami.CardID` is a string type, but its values are numeric IDs.
  For any ordering or comparison logic (sorting, range checks, thresholds),
  use the `Int()` method instead of comparing the strings lexicographically.
