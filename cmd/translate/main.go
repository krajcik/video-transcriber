package main

import (
	"flag"
	"os"

	"github.com/go-pkgz/lgr"

	"assemblyai-transcriber/internal/config"
	"assemblyai-transcriber/internal/database"
	"assemblyai-transcriber/internal/openrouter"
	"assemblyai-transcriber/internal/translation"
)

func main() {
	lgr.Setup()
	// Parse command line arguments
	idFlag := flag.Int64("id", 0, "Transcription ID to translate")
	langFlag := flag.String("lang", "ru", "Target language (e.g. 'ru')")
	allFlag := flag.Bool("all", false, "Translate all untranslated transcriptions")
	flag.Parse()

	// Validate arguments
	if *idFlag == 0 && !*allFlag {
		lgr.Printf("Must specify either --id or --all")
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		lgr.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		lgr.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	// Initialize OpenRouter client
	openrouterClient := openrouter.New(cfg.OpenRouterAPIKey)

	// Create translation service
	translationService := translation.New(db, openrouterClient)

	// Execute translation based on flags
	if *idFlag > 0 {
		translateSingle(*idFlag, *langFlag, translationService)
	} else if *allFlag {
		translateAll(*langFlag, translationService)
	}
}

// translateSingle translates a single transcription
func translateSingle(id int64, lang string, service *translation.Service) {
	lgr.Printf("Translating transcription ID %d to %s...", id, lang)

	err := service.ProcessTranscription(id)
	if err != nil {
		lgr.Fatalf("Translation error: %v", err)
	}

	lgr.Printf("Translation completed and saved to database")
}

// translateAll translates all untranslated transcriptions
func translateAll(lang string, _ *translation.Service) {
	lgr.Printf("Finding untranslated transcriptions for %s translation...", lang)

	// TODO: Implement logic to find and translate all untranslated transcriptions
	lgr.Printf("Batch translation functionality will be implemented in next version")
}
