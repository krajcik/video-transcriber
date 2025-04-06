package translation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"assemblyai-transcriber/internal/database"
	"assemblyai-transcriber/internal/openrouter"
	"assemblyai-transcriber/internal/terms"
)

// Service manages the translation workflow
type Service struct {
	db          *database.DB
	openrouter  *openrouter.Client
	termManager *terms.TermManager
}

// New creates a new translation service
func New(db *database.DB, openrouter *openrouter.Client) *Service {
	return &Service{
		db:          db,
		openrouter:  openrouter,
		termManager: terms.New(),
	}
}

// ProcessTranscription analyzes and translates a transcription
func (s *Service) ProcessTranscription(transcriptionID int64) error {
	// Get the transcription text
	text, err := s.db.GetTranscription(transcriptionID)
	if err != nil {
		return fmt.Errorf("error retrieving transcription: %w", err)
	}

	// Analyze terms
	if err := s.analyzeTerms(text); err != nil {
		return fmt.Errorf("error analyzing terms: %w", err)
	}

	// Process terms interactively
	if err := s.termManager.ProcessTermsInteractive(); err != nil {
		return fmt.Errorf("error processing terms: %w", err)
	}

	// Save terms to database
	if err := s.saveTerms(); err != nil {
		return fmt.Errorf("error saving terms: %w", err)
	}

	// Translate text
	translatedText, err := s.translateText(text)
	if err != nil {
		return fmt.Errorf("error translating text: %w", err)
	}

	// Save translation to database
	if err := s.db.SaveTranslation(transcriptionID, translatedText); err != nil {
		return fmt.Errorf("error saving translation: %w", err)
	}

	fmt.Println("Translation process completed successfully!")
	return nil
}

// analyzeTerms analyzes text for terms that should not be translated
func (s *Service) analyzeTerms(text string) error {
	fmt.Println("Analyzing text for specialized terms...")

	// Analyze text with OpenRouter
	analysis, err := s.openrouter.AnalyzeTerms(text)
	if err != nil {
		return fmt.Errorf("error analyzing terms: %w", err)
	}

	// Convert to Term objects
	var termsList []*terms.Term
	for _, t := range analysis.Terms {
		termsList = append(termsList, &terms.Term{
			Term:        t.Term,
			Description: t.Description,
			Context:     t.Context,
			Keep:        true, // Default to keeping
		})
	}

	// Add to term manager
	s.termManager.AddTerms(termsList)
	return nil
}

// saveTerms saves the terms to the database
func (s *Service) saveTerms() error {
	allTerms := s.termManager.GetAllTerms()
	for _, term := range allTerms {
		if term.Keep {
			if err := s.db.SaveTerm(term.Term, term.Description); err != nil {
				return fmt.Errorf("error saving term '%s': %w", term.Term, err)
			}
		}
	}
	return nil
}

// translateText translates the text using OpenRouter
func (s *Service) translateText(text string) (string, error) {
	fmt.Println("Translating text...")

	// Get list of untranslatable terms
	untranslatableTerms := s.termManager.GetUntranslatableTerms()

	// Translate text
	translatedText, err := s.openrouter.TranslateText(text, untranslatableTerms)
	if err != nil {
		return "", fmt.Errorf("error translating text: %w", err)
	}

	return translatedText, nil
}

// SaveTranscriptionToFile saves a transcription to a file
func (s *Service) SaveTranscriptionToFile(transcriptionID int64, outputPath string) error {
	// Get the transcription text
	text, err := s.db.GetTranscription(transcriptionID)
	if err != nil {
		return fmt.Errorf("error retrieving transcription: %w", err)
	}

	// Create output directory if needed
	outputDir := filepath.Dir(outputPath)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return fmt.Errorf("error creating output directory: %w", err)
		}
	}

	// Write to file
	if err := os.WriteFile(outputPath, []byte(text), 0644); err != nil {
		return fmt.Errorf("error writing transcription to file: %w", err)
	}

	fmt.Printf("Transcription saved to %s\n", outputPath)
	return nil
}

// SaveTranslationToFile saves a translation to a file
func (s *Service) SaveTranslationToFile(transcriptionID int64, outputPath string) error {
	// Get the translation text
	text, err := s.db.GetTranslation(transcriptionID)
	if err != nil {
		return fmt.Errorf("error retrieving translation: %w", err)
	}

	// Create output directory if needed
	outputDir := filepath.Dir(outputPath)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return fmt.Errorf("error creating output directory: %w", err)
		}
	}

	// Write to file
	if err := os.WriteFile(outputPath, []byte(text), 0644); err != nil {
		return fmt.Errorf("error writing translation to file: %w", err)
	}

	fmt.Printf("Translation saved to %s\n", outputPath)
	return nil
}

// GenerateFileName generates an output filename based on the input file
func GenerateFileName(inputFile, suffix, ext string) string {
	base := filepath.Base(inputFile)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	return fmt.Sprintf("%s_%s%s", name, suffix, ext)
}
