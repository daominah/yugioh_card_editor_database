-- Cards: TCG / OCG / Master Duel
-- Boolean columns use INTEGER (0/1): SQLite has no BOOL type.
-- Array columns use TEXT storing JSON: SQLite has no JSONB type.
CREATE TABLE IF NOT EXISTS cards
(
    card_id         INTEGER PRIMARY KEY,           -- Konami card ID, e.g. 4007
    -- card_name_en: EN name as printed on the JA card page <h1> (the bare <span>
    -- next to the kanji title). Only the "ja" crawl pass writes this column,
    -- as a convenience so DB viewers can identify rows without joining card_texts.
    -- Authoritative localized EN text lives in card_texts where lang='en'.
    card_name_en    TEXT    NOT NULL DEFAULT '',
    card_type       TEXT    NOT NULL DEFAULT '',   -- "Monster", "Spell", "Trap"
    card_subtype    TEXT    NOT NULL DEFAULT '',   -- "MonsterNormal", "SpellField", etc.
    attribute       TEXT    NOT NULL DEFAULT '',   -- "LIGHT", "DARK", etc.
    monster_type    TEXT    NOT NULL DEFAULT '',
    level_rank_link INTEGER NOT NULL DEFAULT 0,
    atk             INTEGER NOT NULL DEFAULT 0,
    atk_str         TEXT    NOT NULL DEFAULT '',   -- "?" for some cards
    def             INTEGER NOT NULL DEFAULT 0,
    def_str         TEXT    NOT NULL DEFAULT '',   -- "?" for some cards
    abilities       TEXT    NOT NULL DEFAULT '[]', -- JSON array, e.g. ["Tuner","Flip"]
    link_arrows     TEXT    NOT NULL DEFAULT '[]', -- JSON array, e.g. ["Up","DownLeft"]
    is_pendulum             INTEGER NOT NULL DEFAULT 0,
    pendulum_scale          INTEGER NOT NULL DEFAULT 0,
    is_non_effect           INTEGER NOT NULL DEFAULT 0,
    -- Special Summon-only ("Nomi" / "Semi-Nomi"): cannot be Normal Summoned or Set;
    -- only Special Summoned via card-text conditions. Detected from the JA
    -- "特殊召喚" token in the species block; Konami's EN page omits this flag.
    is_special_summon_only  INTEGER NOT NULL DEFAULT 0,
    year            TEXT    NOT NULL DEFAULT '',   -- first TCG release year, e.g. "2002"
    creator         TEXT    NOT NULL DEFAULT ''
);

-- 8-digit passwords printed on physical cards. Sourced from ygocdb.com (third-
-- party); Konami's official DB does not expose this field. Kept in its own
-- table so the canonical cards row stays free of YGOCDB's data lifecycle.
-- Join with cards.card_id when password is needed; LEFT JOIN if a card may
-- not yet have a YGOCDB entry. card_name is a denormalized convenience
-- (en_name with jp_name fallback) so the row is human-readable on its own.
CREATE TABLE IF NOT EXISTS card_passwords
(
    card_id   INTEGER PRIMARY KEY, -- Konami cardID, joins with cards.card_id (no FK)
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
    lang              TEXT    NOT NULL,            -- "en", "ja", "ko"
    name              TEXT    NOT NULL DEFAULT '', -- official name: EN text, kanji for JA, Hangul for KO
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
    set_code       TEXT PRIMARY KEY,         -- abbreviation, e.g. "LOB", "LOCH", "QCAC"
    game_version   TEXT NOT NULL DEFAULT '', -- "OCG", "TCG", "RushDuel"
    release_date   TEXT NOT NULL DEFAULT '', -- "YYYY-MM-DD"
    name_en        TEXT NOT NULL DEFAULT '', -- set name from EN card pages
    name_ja        TEXT NOT NULL DEFAULT '', -- set name from JA card pages
    name_ko        TEXT NOT NULL DEFAULT '', -- set name from KO card pages
    name_yugipedia TEXT NOT NULL DEFAULT ''  -- set name from Yugipedia
);

-- All prints of every card across all locales
CREATE TABLE IF NOT EXISTS set_cards
(
    card_set_code TEXT PRIMARY KEY,            -- full card number, e.g. "LOCH-JP077", "DBLE-KRS03"
    set_code      TEXT    NOT NULL REFERENCES sets (set_code),
    card_id       INTEGER NOT NULL DEFAULT 0,  -- 0 if card not yet resolved
    rarity_code   TEXT    NOT NULL DEFAULT '', -- short code, e.g. "UL", "SE"
    rarity_name   TEXT    NOT NULL DEFAULT '', -- full name, e.g. "Ultimate Rare"
    release_date  TEXT    NOT NULL DEFAULT ''  -- "YYYY-MM-DD"
);

-- Monster type enum to localized display text (TCG / OCG / Master Duel)
CREATE TABLE IF NOT EXISTS monster_types
(
    monster_type TEXT PRIMARY KEY,         -- canonical enum, e.g. "Dragon"
    text_en      TEXT NOT NULL DEFAULT '', -- "Dragon"
    text_ja      TEXT NOT NULL DEFAULT '', -- "ドラゴン族"
    text_ko      TEXT NOT NULL DEFAULT ''  -- "드래곤족"
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
    text_en   TEXT NOT NULL DEFAULT '', -- "LIGHT"
    text_ja   TEXT NOT NULL DEFAULT '', -- "光属性"
    text_ko   TEXT NOT NULL DEFAULT ''  -- "빛"
);

-- Monster attribute enum to localized display text (Rush Duel / Duel Links)
CREATE TABLE IF NOT EXISTS monster_attributes_rush
(
    attribute TEXT PRIMARY KEY, -- canonical enum, e.g. "LIGHT"
    text_ja   TEXT NOT NULL DEFAULT '',
    text_ko   TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_set_cards_card_id ON set_cards (card_id);
CREATE INDEX IF NOT EXISTS idx_set_cards_set_code ON set_cards (set_code);
CREATE INDEX IF NOT EXISTS idx_card_texts_card_id ON card_texts (card_id);
CREATE INDEX IF NOT EXISTS idx_sets_game_version ON sets (game_version);
