package translation

import (
	"fmt"
	"testing"

	"assemblyai-transcriber/internal/database"
	"assemblyai-transcriber/internal/openrouter"

	"github.com/stretchr/testify/require"
)

func TestGenerateFileName(t *testing.T) {
	t.Run("basic case", func(t *testing.T) {
		result := GenerateFileName("test.txt", "ru", ".txt")
		require.Equal(t, "test_ru.txt", result)
	})

	t.Run("with version suffix", func(t *testing.T) {
		result := GenerateFileName("test.txt", "ru", "_v1.txt")
		require.Equal(t, "test_ru_v1.txt", result)
	})

	t.Run("empty extension", func(t *testing.T) {
		result := GenerateFileName("test", "en", "")
		require.Equal(t, "test_en", result)
	})

	t.Run("complex case", func(t *testing.T) {
		result := GenerateFileName("file.name.with.dots.txt", "fr", "_v2.json")
		require.Equal(t, "file.name.with.dots_fr_v2.json", result)
	})
}

type mockDB struct {
	getTranscriptionFunc func(int64) (string, error)
	getTranslationFunc   func(int64) (string, error)
	saveTermFunc         func(string, string) error
	saveTranslationFunc  func(int64, string) error
}

func (m *mockDB) GetTranscription(id int64) (string, error) {
	return m.getTranscriptionFunc(id)
}

func (m *mockDB) GetTranslation(id int64) (string, error) {
	return m.getTranslationFunc(id)
}

func (m *mockDB) SaveTerm(term, desc string) error {
	return m.saveTermFunc(term, desc)
}

func (m *mockDB) SaveTranslation(id int64, text string) error {
	return m.saveTranslationFunc(id, text)
}

type mockOpenRouter struct {
	analyzeTermsFunc  func(string) (*openrouter.TermAnalysis, error)
	translateTextFunc func(string, []string) (string, error)
}

func (m *mockOpenRouter) AnalyzeTerms(text string) (*openrouter.TermAnalysis, error) {
	return m.analyzeTermsFunc(text)
}

func (m *mockOpenRouter) TranslateText(text string, terms []string) (string, error) {
	return m.translateTextFunc(text, terms)
}

func TestNew(t *testing.T) {
	db := &database.DB{}
	or := &openrouter.Client{}
	tr := New(db, or)
	require.NotNil(t, tr)
	require.Equal(t, db, tr.db)
	require.Equal(t, or, tr.openrouter)
	require.NotNil(t, tr.termManager)
}

func TestProcessTranscription_Success(t *testing.T) {
	db := &mockDB{
		getTranscriptionFunc: func(id int64) (string, error) {
			return "test text", nil
		},
		getTranslationFunc: func(id int64) (string, error) {
			return "translated text", nil
		},
		saveTermFunc: func(term, desc string) error {
			return nil
		},
		saveTranslationFunc: func(id int64, text string) error {
			return nil
		},
	}

	or := &mockOpenRouter{
		analyzeTermsFunc: func(text string) (*openrouter.TermAnalysis, error) {
			return &openrouter.TermAnalysis{}, nil
		},
		translateTextFunc: func(text string, terms []string) (string, error) {
			return "translated text", nil
		},
	}

	tr := New(db, or)
	err := tr.ProcessTranscription(1)
	require.NoError(t, err)
}

func TestProcessTranscription_GetTranscriptionError(t *testing.T) {
	db := &mockDB{
		getTranscriptionFunc: func(id int64) (string, error) {
			return "", fmt.Errorf("db error")
		},
		getTranslationFunc: func(id int64) (string, error) {
			return "", nil
		},
	}

	or := &mockOpenRouter{}
	tr := New(db, or)
	err := tr.ProcessTranscription(1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "error retrieving transcription")
}
