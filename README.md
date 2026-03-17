# YuGiOh card editor

This web app can load Konami card text then you can insert your card art, so you
have a high resolution card. Or you can create a completely new custom card from
scratch. Output is a PNG image that has a resolution of 1180x1720 or 590x860
(based on real card size is 59mm x 86mm).

Inspired by [DuelingBook](https://www.duelingbook.com/) custom card maker.

Well tested on Firefox Linux.

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
runs daily to crawl the Konami database (`cmd/crawl_konami_db`), enrich it with
card passwords and set info (`cmd/add_card_password`), then push the updated
`konami_db_en.js` to the GitHub Pages repo automatically.

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
Alternatively, remove that flag and run `go run cmd/main_yugioh_card_editor/main.go`
to serve on `http://localhost:20808`.

## All cards list table

Final result is file [yugioh_cards.pdf](internal/core/yugioh_cards.pdf).
All cards shown as a table (without card effect text).

The [Google Drive file yugioh_cards.gsheet](
https://docs.google.com/spreadsheets/d/1EzqMmwNq6jc_4JbxjxvjK8EdCHBTTyal248kmG2BuZ0/edit?usp=sharing).

Steps to generate this file:

1. Run `cmd/add_card_password` to get file `yugioh_cards.csv`.
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
