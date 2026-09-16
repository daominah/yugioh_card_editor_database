-- Cards: TCG / OCG / Master Duel
-- Boolean columns use INTEGER (0/1): SQLite has no BOOL type.
-- Array columns use TEXT storing JSON: SQLite has no JSONB type.
CREATE TABLE IF NOT EXISTS cards
(
    card_id                INTEGER PRIMARY KEY,           -- Konami card ID, e.g. 4007
    -- card_name_en: EN name as printed on the JA card page <h1> (the bare <span>
    -- next to the kanji title). Only the "ja" crawl pass writes this column,
    -- as a convenience so DB viewers can identify rows without joining card_texts.
    -- Authoritative localized EN text lives in card_texts where lang='en'.
    card_name_en           TEXT    NOT NULL DEFAULT '',
    card_type              TEXT    NOT NULL DEFAULT '',   -- "Monster", "Spell", "Trap"
    card_subtype           TEXT    NOT NULL DEFAULT '',   -- "MonsterNormal", "SpellField", etc.
    attribute              TEXT    NOT NULL DEFAULT '',   -- "LIGHT", "DARK", etc.
    monster_type           TEXT    NOT NULL DEFAULT '',
    level_rank_link        INTEGER NOT NULL DEFAULT 0,
    atk                    INTEGER NOT NULL DEFAULT 0,
    atk_str                TEXT    NOT NULL DEFAULT '',   -- "?" for some cards
    def                    INTEGER NOT NULL DEFAULT 0,
    def_str                TEXT    NOT NULL DEFAULT '',   -- "?" for some cards
    abilities              TEXT    NOT NULL DEFAULT '[]', -- JSON array, e.g. ["Tuner","Flip"]
    link_arrows            TEXT    NOT NULL DEFAULT '[]', -- JSON array, e.g. ["Up","DownLeft"]
    is_pendulum            INTEGER NOT NULL DEFAULT 0,
    pendulum_scale         INTEGER NOT NULL DEFAULT 0,
    is_non_effect          INTEGER NOT NULL DEFAULT 0,
    -- Special Summon-only ("Nomi" / "Semi-Nomi"): cannot be Normal Summoned or Set;
    -- only Special Summoned via card-text conditions. Detected from the JA
    -- "特殊召喚" token in the species block; Konami's EN page omits this flag.
    is_special_summon_only INTEGER NOT NULL DEFAULT 0,
    year                   TEXT    NOT NULL DEFAULT '',   -- first TCG release year, e.g. "2002"
    creator                TEXT    NOT NULL DEFAULT ''
);

-- 8-digit passwords printed on physical cards. Sourced from ygocdb.com (third-party);
-- Konami's official DB does not expose this field. Kept in its own
-- table so the canonical cards row stays free of YGOCDB's data lifecycle.
-- Join with cards.card_id when password is needed; LEFT JOIN if a card may
-- not yet have a YGOCDB entry. card_name is a denormalized convenience
-- (en_name with jp_name fallback) so the row is human-readable on its own.
CREATE TABLE IF NOT EXISTS card_passwords
(
    card_id   INTEGER PRIMARY KEY,      -- Konami cardID, joins with cards.card_id (no FK)
    password  TEXT NOT NULL DEFAULT '', -- 8-digit, zero-padded, e.g. "08233522"
    card_name TEXT NOT NULL DEFAULT ''  -- en_name with jp_name fallback (YGOCDB at crawl time)
);

-- Cards: Rush Duel / Duel Links
-- Boolean columns use INTEGER (0/1): SQLite has no BOOL type.
-- Array columns use TEXT storing JSON: SQLite has no JSONB type.
CREATE TABLE IF NOT EXISTS cards_rush
(
    card_id         INTEGER PRIMARY KEY,           -- Konami Rush Duel card ID, e.g. 15150
    card_type       TEXT    NOT NULL DEFAULT '',
    card_subtype    TEXT    NOT NULL DEFAULT '',
    attribute       TEXT    NOT NULL DEFAULT '',
    monster_type    TEXT    NOT NULL DEFAULT '',
    level_rank_link INTEGER NOT NULL DEFAULT 0,
    atk             INTEGER NOT NULL DEFAULT 0,
    atk_str         TEXT    NOT NULL DEFAULT '',
    def             INTEGER NOT NULL DEFAULT 0,
    def_str         TEXT    NOT NULL DEFAULT '',
    abilities       TEXT    NOT NULL DEFAULT '[]', -- JSON array
    is_non_effect   INTEGER NOT NULL DEFAULT 0,
    rush_is_legend  INTEGER NOT NULL DEFAULT 0,
    rush_max_atk    TEXT    NOT NULL DEFAULT ''    -- MAXIMUM ATK, non-empty for Maximum Monsters only
);

-- Card text per language (shared between cards and cards_rush; IDs do not collide)
CREATE TABLE IF NOT EXISTS card_texts
(
    card_id           INTEGER NOT NULL,
    lang              TEXT    NOT NULL,            -- "ja", "ko", "en"
    name              TEXT    NOT NULL DEFAULT '', -- official name: kanji for JA, Hangul for KO, English text for EN
    name_katakana     TEXT    NOT NULL DEFAULT '', -- katakana pronunciation for JA; empty for EN and KO
    effect            TEXT    NOT NULL DEFAULT '',
    pendulum_effect   TEXT    NOT NULL DEFAULT '',
    attribute_text    TEXT    NOT NULL DEFAULT '', -- localized attribute: "LIGHT" / "光属性" / "빛"
    -- localized monster type: "Dragon" / "ドラゴン族" / "드래곤족" / Rush "竜族".
    -- Sourced verbatim from the first slash-separated token of <p class="species">,
    -- only on monster pages (detected by presence of ATK/DEF item_box_value spans).
    -- Empty for spells, traps, and any non-monster row.
    monster_type_text TEXT    NOT NULL DEFAULT '',
    PRIMARY KEY (card_id, lang)
);

-- Sets (OCG / TCG / Rush Duel)
CREATE TABLE IF NOT EXISTS sets
(
    set_code     TEXT PRIMARY KEY,         -- abbreviation, e.g. "LOB", "LOCH", "QCAC"
    game_version TEXT NOT NULL DEFAULT '', -- "OCG", "TCG", "RushDuel"
    release_date TEXT NOT NULL DEFAULT '', -- "YYYY-MM-DD"
    name_ja      TEXT NOT NULL DEFAULT '', -- set name from JA card pages
    name_ko      TEXT NOT NULL DEFAULT '', -- set name from KO card pages
    name_en      TEXT NOT NULL DEFAULT ''  -- set name from EN card pages
);

-- All prints of every card across all locales.
-- The same card_set_code can appear with multiple rarities (e.g. a card printed
-- as both Ultra Rare and Secret Rare in the same set), so the primary key is
-- composite: (card_set_code, rarity_code).
-- rarity_name is the localized name from the source Konami page; the language
-- is recoverable from card_set_code's region suffix (JP→ja, EN→en, KR→ko, etc.).
-- A given (card_set_code, rarity_code) is normally written by exactly one
-- crawl pass; for the rare promo/Asia codes that appear on multiple locale
-- pages, first-write-wins (crawl order ja→en→ko, so JA — usually the earliest
-- release — wins). UpsertSetCards relies on the ON CONFLICT clause below
-- and does not log conflicts; run cmd/read-yugiohdb-stats to audit.
CREATE TABLE IF NOT EXISTS set_cards
(
    card_set_code TEXT    NOT NULL,            -- full card number, e.g. "LOCH-JP077", "DBLE-KRS03"
    set_code      TEXT    NOT NULL,
    card_id       INTEGER NOT NULL DEFAULT 0,  -- 0 if card not yet resolved
    rarity_code   TEXT    NOT NULL DEFAULT '', -- short code, e.g. "UL", "SE"
    rarity_name   TEXT    NOT NULL DEFAULT '', -- localized name from source page, e.g. "Ultimate Rare", "ウルトラレア仕様", "울트라 레어"
    release_date  TEXT    NOT NULL DEFAULT '', -- "YYYY-MM-DD"
    PRIMARY KEY (card_set_code, rarity_code)
);

-- Canonical rarity lookup, one row per rarity_code. Per-language columns are
-- needed because rarity_name varies across game languages (unlike attribute
-- or monster_type, which share canonical enum codes across languages).
-- Seeded by cmd/aggregate-type-attr-rarity after each crawl: it reads
-- set_cards rows (already populated with localized names by the crawl
-- passes), infers the language from each row's card_set_code region suffix,
-- groups by (rarity_code, lang), and writes the most-frequent variant into
-- the matching rarity_name_{ja,ko,en}. No HTML is re-scraped here.
CREATE TABLE IF NOT EXISTS rarities
(
    rarity_code    TEXT PRIMARY KEY,            -- short code, e.g. "UL", "SE"
    -- alias_code is the peer code for the same logical rarity under a different
    -- locale convention. Pairs are symmetric: if C→N then N→C. Hard-coded in
    -- cmd/aggregate-type-attr-rarity; empty string means no known alias.
    alias_code     TEXT    NOT NULL DEFAULT '',
    rarity_name_ja TEXT    NOT NULL DEFAULT '', -- canonical JA name, e.g. "ウルトラレア"
    rarity_name_ko TEXT    NOT NULL DEFAULT '', -- canonical KO name, e.g. "울트라 레어"
    rarity_name_en TEXT    NOT NULL DEFAULT '', -- canonical EN name, e.g. "Ultimate Rare"
    -- distinct card_id count from set_cards; reset on each aggregation run.
    -- alias pairs (e.g. C and N) should show similar counts as a sanity check.
    count_cards    INTEGER NOT NULL DEFAULT 0
);

-- Monster type enum to localized display text (TCG / OCG / Master Duel)
CREATE TABLE IF NOT EXISTS monster_types
(
    monster_type TEXT PRIMARY KEY,         -- canonical enum, e.g. "Dragon"
    text_ja      TEXT NOT NULL DEFAULT '', -- "ドラゴン族"
    text_ko      TEXT NOT NULL DEFAULT '', -- "드래곤족"
    text_en      TEXT NOT NULL DEFAULT ''  -- "Dragon"
);

-- Monster type enum to localized display text (Rush Duel / Duel Links)
CREATE TABLE IF NOT EXISTS monster_types_rush
(
    monster_type TEXT PRIMARY KEY,
    text_ja      TEXT NOT NULL DEFAULT '',
    text_ko      TEXT NOT NULL DEFAULT ''
);

-- Monster attribute enum to localized display text (TCG / OCG / Master Duel)
CREATE TABLE IF NOT EXISTS monster_attributes
(
    attribute TEXT PRIMARY KEY,         -- canonical enum, e.g. "LIGHT"
    text_ja   TEXT NOT NULL DEFAULT '', -- "光属性"
    text_ko   TEXT NOT NULL DEFAULT '', -- "빛"
    text_en   TEXT NOT NULL DEFAULT ''  -- "LIGHT"
);

-- Monster attribute enum to localized display text (Rush Duel / Duel Links)
CREATE TABLE IF NOT EXISTS monster_attributes_rush
(
    attribute TEXT PRIMARY KEY, -- canonical enum, e.g. "LIGHT"
    text_ja   TEXT NOT NULL DEFAULT '',
    text_ko   TEXT NOT NULL DEFAULT ''
);
