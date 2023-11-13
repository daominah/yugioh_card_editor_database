# YuGiOh card editor

This client web app can make YuGiOh custom card, output is a PNG image that has
a resolution of 1180x1720 (because real YuGiOh card size is 59mm x 86mm, this
resolution uses 20px to represent 1mm).

Inspired by [DuelingBook](https://www.duelingbook.com/) custom card maker.

## Feature

* High quality, natural resolution ratio output image.
* Can work without network.
* Auto choose font size for card effect, the font can be twitched manually for
  optimizing text space too.
* Crawled Cards data from Konami card database, that can be searched easily
  to be used on the editor.
* TODO: save default card template, export/import card as JSON, save/load user  
  cards gallery on server database.

## Usage

The web app can be run by locally open file `web/index.html`.

Alternatively, you can `go run cmd/main_yugioh_card_editor/main.go` then go to
<http://localhost:20808/>. Or if you prefer Docker, run `build_run.sh` to build
and run this app by Docker. 
