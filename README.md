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

The web app can be run by locally open file `web/index.html`.

Alternatively, you can `go run cmd/main_yugioh_card_editor/main.go` then go to
<http://localhost:20808/>. Or if you prefer Docker, run `build_run.sh` to build
and run this app by Docker.

## Card data sources

* Konami official [card database](https://www.db.yugioh-card.com/yugiohdb/card_search.action?ope=2&cid=4007&request_locale=en).
* Additional data <https://ygocdb.com/api/v0/cards.zip> for card Password.
  Password is 8-digit number that usually printed on the bottom left of the card.
  Example: "Blue-Eyes White Dragon" has Password 89631139, and ID 4007 in Konami database.

## All cards list table

Final result is file [yugioh_cards.pdf](internal\core\yugioh_cards.pdf).
All cards shown as a table (without card effect text)

Steps to generate this file:

1. Run `cmd/add_card_password` to get file `yugioh_cards.csv`.
2. Open it as XLSX file in LibreOffice Calc. Format rows color with
   `Format`: `Conditional`, using `Formula is`:

  ```excel
  AND($B2="Monster", $C2="Normal", ISEVEN(ROW()))  // Yellow
  AND($B2="Monster", $C2="Normal", ISODD(ROW()))   // Light Yellow
  AND($B2="Monster", $C2<>"Normal", ISEVEN(ROW())) // Orange
  AND($B2="Monster", $C2<>"Normal", ISODD(ROW()))  // Light Orange
  AND($B2="Spell", ISEVEN(ROW()))                  // Green
  AND($B2="Spell", ISODD(ROW()))                   // Light Green
  AND($B2="Trap", ISEVEN(ROW()))                   // Purple
  AND($B2="Trap", ISODD(ROW()))                    // Light Purple
  ```

3. Upload to Google Drive to Download as PDF
   (LibreOffice hangs when exporting to PDF, probably because of file too large),
   change page size to Height 19.9", Width 19", so 100 rows fit in 1 page.
   The online [file](https://docs.google.com/spreadsheets/d/1EzqMmwNq6jc_4JbxjxvjK8EdCHBTTyal248kmG2BuZ0/edit?usp=sharing).
