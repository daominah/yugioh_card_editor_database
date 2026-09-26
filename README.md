# YuGiOh card editor

This web app can load Konami card text then you can insert your card art, so you
have a high resolution card. Or you can create a completely new custom card from
scratch. Output is a PNG image that has a resolution of 1180x1720 or 590x860
(based on real card size is 59mm x 86mm).

Inspired by [DuelingBook](https://www.duelingbook.com/) custom card maker.

Well tested on Firefox.

## Feature

* High quality, natural resolution ratio output image.
* Can work without network.
* Auto choose font size for card effect, the font can be twitched manually for
  optimizing text space too.
* Crawled cards data from Konami, that can be searched easily by card name to
  be used on the editor.

## Usage

The web app is hosted on [GitHub Page daominah.github.io](https://daominah.github.io/).
To update the site, run `copy_to_daominah.github.io.sh` to sync the `web/`
directory to the local `daominah.github.io` repo, then push the changes.

You can also open `web/index.html` locally in a browser.

## Card data sources

* Konami official [card database](https://www.db.yugioh-card.com/yugiohdb/card_search.action?ope=2&cid=4007&request_locale=en).
* Additional data <https://ygocdb.com/api/v0/cards.zip> for card Password.
  Password is 8-digit number that usually printed on the bottom left of the card.
  Example: "Blue-Eyes White Dragon" has Password 89631139, and ID 4007 in Konami database.

A [GitHub Actions workflow](.github/workflows/update_cards_database_from_konami.yml)
runs daily to crawl the Konami database (`cmd/crawl-konami-db`), enrich it with
card passwords and set info (`cmd/add-card-password`), then push the updated
`konami_db_en.js` to the GitHub Pages repo automatically.

## Full card database (SQLite)

[`data/yugioh.db`](data/yugioh.db) is a self-contained SQLite snapshot covering
Standard (TCG / OCG) and Rush Duel cards across en, ja, ko locales, plus every
Konami set and print. Schema lives in
[`init_schema.sql`](pkg/driver/sqlite/init_schema.sql).

Contents from the 2026-09 crawl:

| Table                                           | Rows    | Notes                                                                                                                       |
|-------------------------------------------------|---------|-----------------------------------------------------------------------------------------------------------------------------|
| `cards`                                         | 14,183  | Standard cards (TCG / OCG / Master Duel)                                                                                    |
| `cards_rush`                                    | 3,114   | Rush Duel / Duel Links cards                                                                                                |
| `card_texts`                                    | 47,492  | per-locale name and effect (ja: 17,297 / ko: 16,415 / en: 13,780)                                                           |
| `card_passwords`                                | 14,255  | 8-digit passwords from ygocdb.com (Konami's DB does not expose this)                                                        |
| `sets`                                          | 1,553   | every Konami set across all locales                                                                                         |
| `set_cards`                                     | 123,347 | every print of every card; PK is composite (`card_set_code`, `rarity_code`) so one print can be listed at multiple rarities |
| `rarities`                                      | 49      | canonical rarity_code → localized rarity name lookup, seeded by `cmd/aggregate-type-attr-rarity`                            |
| `monster_types`, `monster_types_rush`           | 26 + 29 | enum to localized text (e.g. Dragon → ドラゴン族 / 드래곤족)                                                                         |
| `monster_attributes`, `monster_attributes_rush` | 7 + 6   | enum to localized text (e.g. LIGHT → 光属性 / 빛)                                                                               |

Of the 14,183 Standard cards, 99.7% have a YGOCDB password match. Rush cards
do not, since YGOCDB only tracks Standard. 342 cards are flagged as
Special-Summon-only (Nomi / Semi-Nomi monsters).

To refresh:

```bash
go run cmd/crawl-konami-db-full/crawl_konami_db_full.go
go run cmd/aggregate-type-attr-rarity/aggregate_type_attr_rarity.go
```

The crawler caches every Konami HTML page under `data/html_cache/`
(gitignored), so re-runs do not re-hit Konami's servers.

Two read-only commands query the DB without the `sqlite3` CLI:

- [`cmd/read-yugiohdb-stats`](cmd/read-yugiohdb-stats/read_yugiohdb_stats.go)
  prints the table counts above, plus audit checks for fields the parser failed to fill.
  Modify it freely to add one-off queries.
- [`cmd/read-yugiohdb-by-cardid`](cmd/read-yugiohdb-by-cardid/read_yugiohdb_by_cardid.go)
  takes a card ID and dumps every stored field of that card,
  joining `card_texts` for all locales, the password, and every print with its set and rarity:

```bash
go run cmd/read-yugiohdb-by-cardid/read_yugiohdb_by_cardid.go 4007
```

[`cmd/masterduel-card-decoder`](cmd/masterduel-card-decoder/masterduel-card-decoder.md)
ranks the next guess in the Master Duel card decoder game
by how well it splits the remaining candidates.

## Frontend development with Claude Code

[Playwright MCP](https://github.com/microsoft/playwright-mcp)
lets Claude Code edit `web/` code and open the card editor in a real browser.
Both you and Claude can see the rendered card,
making the edit-preview feedback loop smooth.

Setup (requires Node.js 18+):

```bash
claude mcp add --scope project playwright -- \
  npx -y @playwright/mcp@latest --browser firefox --allow-unrestricted-file-access
```

Install Firefox for Playwright (separate from your system Firefox):

```bash
npx -y playwright install firefox
```

On Windows, manually edit `.mcp.json` to change `"command"` to `"cmd"`
and prepend `"/c"` to the `"args"` array (see [.mcp.json](.mcp.json)).

The `--allow-unrestricted-file-access` flag lets the browser open `web/index.html`
directly via `file://` URL without running a server.
Note that this grants access to any file on the machine.
Alternatively, remove that flag and run `go run cmd/main-yugioh-card-editor/main.go`
to serve on `http://localhost:20808`.

## All cards list table

Final result is file [yugioh_cards.pdf](pkg/core/yugioh_cards.pdf).
All cards shown as a table (without card effect text).

The [Google Drive file yugioh_cards.gsheet](
https://docs.google.com/spreadsheets/d/1EzqMmwNq6jc_4JbxjxvjK8EdCHBTTyal248kmG2BuZ0/edit?usp=sharing).

Steps to generate this file:

1. Run `cmd/add-card-password` to get file `yugioh_cards.csv`.
2. Open it as XLSX file in LibreOffice Calc. Format rows color with
   `Format`: `Conditional`, using `Formula is`:

  ```excel
  AND($C2="Monster", $D2="Normal", ISEVEN(ROW()))  // Yellow
  AND($C2="Monster", $D2="Normal", ISODD(ROW()))   // Light Yellow
  AND($C2="Monster", $D2<>"Normal", ISEVEN(ROW())) // Orange
  AND($C2="Monster", $D2<>"Normal", ISODD(ROW()))  // Light Orange
  AND($C2="Spell", ISEVEN(ROW()))                  // Green
  AND($C2="Spell", ISODD(ROW()))                   // Light Green
  AND($C2="Trap", ISEVEN(ROW()))                   // Purple
  AND($C2="Trap", ISODD(ROW()))                    // Light Purple
  ```

3. Upload to Google Drive to Download as PDF
   (LibreOffice hangs when exporting to PDF, probably because of file too large),
   change page size to Height 19.9", Width 19", so 100 rows fit in 1 page.
