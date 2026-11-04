# Fix: Konami card name moved from id to class

## Report

- The workflow "Update cards database from Konami" fails daily since 2026-10-28.
  The last success was 2026-10-27.
- Failing log line: `cardID 4015 ParseKonamiCardHTML returned empty CardName`.
- The crawler stops on an empty card name by design.
- No ticket: solo project.

## Root cause

- Konami changed the card name wrapper
  from `<div id="cardname">` to `<div class="cardname">`.
- Three parsers matched only the id: `pkg/konami/parse.go`,
  `pkg/konami/card_locale.go`, and `pkg/konami/rush_card.go`.
- Both the main site (`yugiohdb`) and Rush Duel (`rushdb`) pages changed.
- Verified by downloading card 4015 (EN) and Rush Duel card 19105 (JA).

### Side finding: English name gone from JA and KO pages

- The new JA and KO pages no longer print the English name
  (the bare `<span>` next to the local name in the `h1`).
  Nothing else on the page carries it: no hidden element, meta tag, or script.
- `cards.card_name_en` was filled only from that span
  during the "ja" pass of `cmd/crawl-konami-db-full`,
  so a full crawl on fresh pages would leave it empty for every card.
- The daily workflow is not affected:
  `cmd/add-card-password` copies `card_name_en` from the `data/yugioh.db` snapshot.
- The "English (Asia)" locale, `request_locale=ae`, is the OCG English edition.
  Its card name matched the EN page on 8 of 9 sampled cards;
  the 9th (11353) has no "ae" page.
  Its set list differs (Asian English prints such as `SDRB-AE001`).

## Plan

- **Refresh fixtures (red)**: re-download every saved page
  in `pkg/konami/test_html_{en,ja,ko}` and `pkg/konami/test_rushduel_html_{ja,ko}`.
  `test_html_escaped_br` stays as a snapshot of the older markup.
- **Fix (green)**: one shared XPath `xpathCardNameBox`
  matches the id or the class `cardname`, used by all three parsers.
- **Test updates for real Konami content changes**:
  reworded EN text of 14356 and 15524,
  `<Y/N>` now kept in 11353,
  a newer first JA print of 4007,
  Miracle Raven (21627) now has JA and KO pages.
- **English name fallback** in `cmd/crawl-konami-db-full`:
  keep the JA span when present (old cached pages),
  else fill from the EN page name.
- **"ae" fallback kept but switched off** (`isFillingNameENFromAsiaEnglish = false`):
  `fillCardNameENFromAsiaEnglish` reads the "ae" page name of the cards still missing one.
  On the 2026-11-04 full crawl, all 404 such cards had no "ae" page, so it filled 0 names.

## Observability

- Daily workflow: the GitHub Actions run succeeds again,
  with the usual duration of about 12 minutes.
  Before: 7 consecutive failures, each about 30 seconds.
- Full crawl: `cmd/read-yugiohdb-stats` reports the count of `cards.card_name_en=''`.

## Full crawl verification

- Copy `data/yugioh.db` to `data/yugioh_before.db`.
- Rename `data/html_cache` to `data/html_cache_backup`,
  so the crawl downloads every page fresh (about 77k pages).
- Run `go run cmd/crawl-konami-db-full/crawl_konami_db_full.go`.
- Compare every table of the two databases by primary key:
  added, removed, and changed rows.
- Delete `data/html_cache_backup` and `data/yugioh_before.db`
  once the comparison shows no data loss.

### Result, 2026-11-04

- The crawl took 45 minutes.
  Then `cmd/aggregate-type-attr-rarity` and `cmd/export-all-cards-csv` ran.
- **No row removed in any table.**
- **No English name lost or renamed.**
  84 cards gained one from the EN page name,
  for example 8767 "Long-Tailed Black Horse".
  Empty `card_name_en` went from 407 to 404:
  81 cards new this month and 323 OCG-only cards, none with an EN page.
- Added: 81 cards, 77 Rush cards, 442 card texts, 11 sets, 1310 set prints.
- Changed by Konami, not by the parser: 100 effect texts (errata),
  25 placeholder passwords replaced by real ones, 8 set names or dates.
- `data/yugioh_cards.csv`, compared without its row number column:
  0 cards removed, the same 84 cards added.

## Checklist

- [x] Step 1-2: understand and reproduce
- [x] Step 3: plan
- [x] Step 4: branch and draft PR
- [x] Step 5: refresh fixtures, confirm red
- [x] Step 6: commit and push red
- [x] Step 7: apply the fix, green
- [x] Step 8: reflect: English name fallback, every-page name test
- [x] Step 9: commit the fix, mark PR ready
- [x] Step 10: self-review
- [ ] Step 11: Linear and Slack (skipped: no ticket)
- [ ] Step 12: review feedback
- [x] Full crawl verification: no data loss
- [ ] Step 13: merge (user-driven)
- [ ] Step 14-15: run the daily workflow manually and confirm the next scheduled run
- [ ] Step 16: announce (skipped: no ticket), not signed off
