-- +goose Up
CREATE TABLE IF NOT EXISTS untranslatable_terms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    term TEXT NOT NULL UNIQUE,
    description TEXT,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS untranslatable_terms;
