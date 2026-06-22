-- Non-PK indexes used for query performance after the bulk crawl is loaded.
-- The crawler deliberately defers these until after the single-transaction
-- bulk load (see DB.CreateIndexes), so each row only touches the four PK
-- B-trees during insert and the index B-trees are built once at the end.
CREATE INDEX IF NOT EXISTS idx_set_cards_card_id ON set_cards (card_id);
CREATE INDEX IF NOT EXISTS idx_set_cards_set_code ON set_cards (set_code);
CREATE INDEX IF NOT EXISTS idx_card_texts_card_id ON card_texts (card_id);
CREATE INDEX IF NOT EXISTS idx_sets_game_version ON sets (game_version);
