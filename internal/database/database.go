package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// DB represents the database connection
type DB struct {
	conn *sql.DB
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Setup initializes the database schema
func (db *DB) Setup() error {
	// create transcriptions table
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS transcriptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_name TEXT NOT NULL,
			transcript_text TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating transcriptions table: %w", err)
	}

	// create untranslatable_terms table
	_, err = db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS untranslatable_terms (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			term TEXT NOT NULL UNIQUE,
			description TEXT,
			added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating untranslatable_terms table: %w", err)
	}

	// create translations table
	_, err = db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS translations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			transcription_id INTEGER,
			translated_text TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (transcription_id) REFERENCES transcriptions(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating translations table: %w", err)
	}

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
	err := db.conn.QueryRow(
		"SELECT transcript_text FROM transcriptions WHERE id = ?",
		id,
	).Scan(&text)
	if err != nil {
		return "", fmt.Errorf("error retrieving transcription: %w", err)
	}

	return text, nil
}

// GetAllTerms retrieves all untranslatable terms
func (db *DB) GetAllTerms() ([]map[string]string, error) {
	rows, err := db.conn.Query("SELECT term, description FROM untranslatable_terms")
	if err != nil {
		return nil, fmt.Errorf("error retrieving terms: %w", err)
	}
	defer rows.Close()

	var terms []map[string]string
	for rows.Next() {
		var term, description string
		if err := rows.Scan(&term, &description); err != nil {
			return nil, fmt.Errorf("error scanning term row: %w", err)
		}
		terms = append(terms, map[string]string{
			"term":        term,
			"description": description,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating term rows: %w", err)
	}

	return terms, nil
}

// GetTranslation retrieves a translation by transcription ID
func (db *DB) GetTranslation(transcriptionID int64) (string, error) {
	var text string
	err := db.conn.QueryRow(
		"SELECT translated_text FROM translations WHERE transcription_id = ?",
		transcriptionID,
	).Scan(&text)
	if err != nil {
		return "", fmt.Errorf("error retrieving translation: %w", err)
	}

	return text, nil
}
