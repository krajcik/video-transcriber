package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// DB represents the database connection
type DB struct {
	conn *sqlx.DB
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	conn, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Setup is deprecated: schema initialization is now handled by goose migrations.
func (db *DB) Setup() error {
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// SaveTranscription saves a transcription to the database
func (db *DB) SaveTranscription(fileName, text string) (int64, error) {
	result, err := db.conn.Exec(
		"INSERT INTO transcriptions (file_name, transcript_text) VALUES (?, ?)",
		fileName, text,
	)
	if err != nil {
		return 0, fmt.Errorf("error saving transcription: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting inserted ID: %w", err)
	}

	return id, nil
}

// SaveTerm saves an untranslatable term to the database
func (db *DB) SaveTerm(term, description string) error {
	_, err := db.conn.Exec(
		"INSERT OR REPLACE INTO untranslatable_terms (term, description) VALUES (?, ?)",
		term, description,
	)
	if err != nil {
		return fmt.Errorf("error saving term: %w", err)
	}

	return nil
}

// SaveTranslation saves a translation to the database
func (db *DB) SaveTranslation(transcriptionID int64, translatedText string) error {
	_, err := db.conn.Exec(
		"INSERT INTO translations (transcription_id, translated_text) VALUES (?, ?)",
		transcriptionID, translatedText,
	)
	if err != nil {
		return fmt.Errorf("error saving translation: %w", err)
	}

	return nil
}

// GetTranscription retrieves a transcription by ID
func (db *DB) GetTranscription(id int64) (string, error) {
	var text string
	err := db.conn.Get(&text,
		"SELECT transcript_text FROM transcriptions WHERE id = ?",
		id,
	)
	if err != nil {
		return "", fmt.Errorf("error retrieving transcription: %w", err)
	}

	return text, nil
}

// GetAllTerms retrieves all untranslatable terms
func (db *DB) GetAllTerms() ([]map[string]string, error) {
	var terms []struct {
		Term        string `db:"term"`
		Description string `db:"description"`
	}
	err := db.conn.Select(&terms, "SELECT term, description FROM untranslatable_terms")
	if err != nil {
		return nil, fmt.Errorf("error retrieving terms: %w", err)
	}

	result := make([]map[string]string, len(terms))
	for i, t := range terms {
		result[i] = map[string]string{
			"term":        t.Term,
			"description": t.Description,
		}
	}

	return result, nil
}

// GetTranslation retrieves a translation by transcription ID
func (db *DB) GetTranslation(transcriptionID int64) (string, error) {
	var text string
	err := db.conn.Get(&text,
		"SELECT translated_text FROM translations WHERE transcription_id = ?",
		transcriptionID,
	)
	if err != nil {
		return "", fmt.Errorf("error retrieving translation: %w", err)
	}

	return text, nil
}
