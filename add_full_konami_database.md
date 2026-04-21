# Add Full Konami Database

See `card_data_collection.md` for the existing pipeline this feature extends.

## Checklist

- [x] Step 1: Understand requirements
- [x] Step 2: Clarify vague points
- [x] Step 4: High-level design (DB schema done: pkg/driver/sqlite/init_schema.sql)
- [ ] Step 5: Detailed implementation plan
- [ ] Step 6: Write red tests
- [ ] Step 7: Draft PR
- [ ] Step 8: Implement
- [ ] Step 9: Document
- [ ] Step 10: Commit and mark PR ready

---

## Context (Step 1)

### Problem

The existing pipeline stores one EN set code per card and no set names or rarities.
Given an order from a Japanese or Korean card shop, each item has:

- A card name that is machine-translated or in Japanese — unreliable
- A set code like `LOCH-JP077` or `DBLE-KRS03` — reliable and unique per print

There is no way to look up a card by its JP/KR set code today.

### Goal

Build a full card database (SQLite) that stores:

- All TCG/OCG/Master Duel cards and Rush Duel/Duel Links cards
- Card text (name, effect) per locale — EN, JA, KO
- All prints of every card per locale: set code, set name, rarity, release date
  (locale and region are coupled: `request_locale=ja` gives JA text + JP set codes,
  `request_locale=ko` gives KO text + KR set codes, etc.)

### First use case

Given a set code from an order sheet (e.g. `LOCH-JP077`, `DBLE-KRS03`),
return the card's full info including official English name.

Example order rows:
```
"The Demon Who Hunts for Sinful Treasures", Super,    LOCH-JP077  ← machine-translated JP name
"The Demon of Sinful Treasure Hunting",     Ultimate, LOCH-JP077  ← same card, different rarity
Clear Wing Synchro Dragon,   Extra Secret Parallel Rare, DBLE-KRS03
Starving Venom Fusion Dragon, Extra Secret Parallel Rare, DBLE-KRS04
```

The same set code can appear multiple times in an order (same card, different rarity/condition).
The set code is the join key: `LOCH-JP077` → card_id → EN name.

### What the web app needs

`web/konami_data/konami_db_en.js` continues to be generated from the DB as an
export step. The web app is not affected.

---

## Clarifications (Step 2)

### Konami DB is the single data source

All new data comes from the same source already used today: `db.yugioh-card.com`.

Each card page already contains everything we need but the current parser discards most of it:

- Card name and effect in the requested locale
- All prints in that locale: date, set code, set name, rarity code, rarity full name
  (the `t_row` structure in `#update_list`, see `konami_card.go:264`)

Each `request_locale` pass yields both the text language and the region's set codes
together — they are not the same concept but always come paired from one crawl pass.
Passwords remain from YGOCDB, same as today.

**JA locale is the primary crawl:** widest card ID coverage (OCG-only cards exist
in JA but not EN); newer cards may have JA text before EN is available.
JA pages also show the EN name when the card has a TCG release.

### HTML cache

Raw HTML saved to `html_cache/{locale}/{card_id}.html` (gitignored) before parsing.
Allows re-parsing during development without re-downloading.
(`crawl_konami_db.go:115` already has a TODO for this.)

### New cmd

A new `cmd/crawl-konami-db-full` handles all locales and both game versions.
`cmd/crawl-konami-db` is kept unchanged as the existing web pipeline.

### Rush Duel / Duel Links

`konami.ParseRushDuelCardHTML` exists but has no cmd yet.
Uses `rushdb` base URL instead of `yugiohdb`. Locales: `ja` and `ko`.
First known card ID: 15150. Card IDs do not collide with TCG/OCG IDs.
Last card ID TBD — same stop strategy: 400 consecutive not-found.

---

## DB design notes (deferred to Step 4)

- SQLite file `yugioh.db` (gitignored)
- Split `cards` (TCG/OCG/Master Duel) and `cards_rush` (Rush Duel/Duel Links) tables
- `game_version` only applies to sets, not cards; values: `OCG`, `TCG`, `RushDuel`
  (matches existing `YuGiOhVersion` type in `pkg/core/set_number.go`)
- Shared `card_texts` and `set_cards` tables (card IDs do not collide)
- `card_texts` stores attribute display text per card per language:
  EN: "LIGHT", JA: "光属性", KO: "빛" — repeated on every card row as scraped
- Separate `attributes` table: `(attribute, text_en, text_ja, text_ko)` —
  reference/lookup, not a FK constraint on `card_texts`
- `card_texts` for JA has two name fields:
  - `name`: kanji name e.g. 青眼の白龍 — official printed name, used by JP shops
  - `name_reading`: katakana reading e.g. ブルーアイズ・ホワイト・ドラゴン — for search
- Sets table stores both Konami set names (per locale, from card pages)
  and Yugipedia set name; full column design deferred
