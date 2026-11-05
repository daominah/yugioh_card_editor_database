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
- To query `data/yugioh.db`, modify and run
  `cmd/read-yugiohdb-stats/read_yugiohdb_stats.go`.
  That file can be freely extended with extra queries and reused between sessions.
- To dump every stored field of one card, run
  `go run cmd/read-yugiohdb-by-cardid/read_yugiohdb_by_cardid.go <card_id>`.
  It joins `cards` (or `cards_rush`), `card_texts`, `card_passwords`,
  `set_cards`, `sets`, and `rarities`.
- Both commands are read-only: they run `SELECT` statements and never write.
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

# Suspicious Empty Fields

- The `--- suspicious empty fields ---` section of `cmd/read-yugiohdb-stats` flags
  possible parser misses. Investigated on the 2026-11-04 crawl:
  every remaining non-zero count is Konami's own data, not a parser miss.
- 404 `cards.card_name_en` are empty: cards with no EN page
  (Konami removed the English name from JA and KO pages around 2026-10-28):
  81 cards new in 2026-10 and 323 OCG-only cards.
- 52 `set_cards.release_date` are empty: Konami's date cell is empty
  (42 `JUMP`, 3 `JMP`, 1 `JMPS`, and `DL1` to `DL5` prizes).
- 28 Pendulum rows with `pendulum_scale=0`: Konami's page shows scale 0.
- 7 non-Link monsters with `level_rank_link=0`:
  real Rank 0 Xyz and Level 0 Synchro monsters.
- 6 ja `card_texts` rows with empty `effect`: Rush Ritual monsters
  whose text box is empty on Konami's page.
- 79 `set_cards` rows without rarity: 76 are prints of MAGNIFICENT MAESTROS (`MAMS`),
  a set released 2026-11-13 whose rarities Konami had not published yet:
  they should fill on the next crawl after release.
  The other 3 (`RD/26IM-KR002`, `RD/26IM-KR003`, `RD/P001-KR001`)
  show no rarity icon on Konami's page.
- Rush Duel empty `card_subtype` (19 rows) was a parser bug, fixed:
  Fusion, Ritual and Union species tokens are katakana or Hangul on Rush pages.
