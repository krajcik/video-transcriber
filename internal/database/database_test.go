package database

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func applyMigrationsForTest(db *DB, t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
	migrationsDir := filepath.Join(projectRoot, "migrations")
	files, err := os.ReadDir(migrationsDir)
	require.NoError(t, err)

	upBlock := regexp.MustCompile(`--\s*\+goose Up([\s\S]*?)(--\s*\+goose|$)`)
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(migrationsDir, file.Name()))
		require.NoError(t, err)
		matches := upBlock.FindStringSubmatch(string(content))
		if len(matches) < 2 {
			continue
		}
		sql := strings.TrimSpace(matches[1])
		if sql == "" {
			continue
		}
		// support for multiple statements in one block
		stmts := strings.Split(sql, ";")
		for _, stmt := range stmts {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			_, err := db.conn.Exec(stmt)
			require.NoError(t, err)
		}
	}
}

func TestDatabase(t *testing.T) {
	t.Run("New and Close", func(t *testing.T) {
		db, err := New(":memory:")
		require.NoError(t, err)
		require.NotNil(t, db)
		require.NoError(t, db.Close())
	})

	t.Run("Setup", func(t *testing.T) {
		db, err := New(":memory:")
		require.NoError(t, err)
		defer db.Close()

		applyMigrationsForTest(db, t)

		// Verify tables exist
		tables := []string{"transcriptions", "untranslatable_terms", "translations"}
		for _, table := range tables {
			_, err := db.conn.Exec("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table)
			require.NoError(t, err, "table %s should exist", table)
		}
	})

	t.Run("Transcription CRUD", func(t *testing.T) {
		db, err := New(":memory:")
		require.NoError(t, err)
		defer db.Close()
		applyMigrationsForTest(db, t)

		// Create
		id, err := db.SaveTranscription("test.mp3", "test transcription")
		require.NoError(t, err)
		require.Positive(t, id)

		// Read
		text, err := db.GetTranscription(id)
		require.NoError(t, err)
		require.Equal(t, "test transcription", text)
	})

	t.Run("Term CRUD", func(t *testing.T) {
		db, err := New(":memory:")
		require.NoError(t, err)
		defer db.Close()
		applyMigrationsForTest(db, t)

		// Create
		err = db.SaveTerm("test term", "test description")
		require.NoError(t, err)

		// Read all
		terms, err := db.GetAllTerms()
		require.NoError(t, err)
		require.Len(t, terms, 1)
		require.Equal(t, "test term", terms[0]["term"])
		require.Equal(t, "test description", terms[0]["description"])
	})

	t.Run("Translation CRUD", func(t *testing.T) {
		db, err := New(":memory:")
		require.NoError(t, err)
		defer db.Close()
		applyMigrationsForTest(db, t)

		// Create transcription first
		transcriptionID, err := db.SaveTranscription("test.mp3", "test transcription")
		require.NoError(t, err)

		// Create translation
		err = db.SaveTranslation(transcriptionID, "test translation")
		require.NoError(t, err)

		// Read
		text, err := db.GetTranslation(transcriptionID)
		require.NoError(t, err)
		require.Equal(t, "test translation", text)
	})
}
