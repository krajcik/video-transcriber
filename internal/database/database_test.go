package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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

		err = db.Setup()
		require.NoError(t, err)

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
		require.NoError(t, db.Setup())

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
		require.NoError(t, db.Setup())

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
		require.NoError(t, db.Setup())

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
