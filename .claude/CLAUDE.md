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

# TODO: Investigate Suspicious Empty Fields

- The `--- suspicious empty fields ---` section of `cmd/read-yugiohdb-stats` flags
  parser misses worth investigating. Counts from the 2026-07 crawl:
  276 cards with empty `card_name_en`, 52 set_cards rows with empty `release_date`,
  28 Pendulum rows with `pendulum_scale=0`, 19 cards_rush rows with empty `card_subtype`,
  7 non-Link monsters with `level_rank_link=0`, 6 ja card_texts rows with empty `effect`,
  and 1 set_cards row missing both `rarity_code` and `rarity_name`.
- MonsterLink rows with empty `link_arrows` used to be the largest group (~462).
  That check now reports 0, so it needs no further investigation.
