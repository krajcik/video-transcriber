package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"assemblyai-transcriber/internal/config"
	"assemblyai-transcriber/internal/database"
	"assemblyai-transcriber/internal/openrouter"
	"assemblyai-transcriber/internal/translation"
)

func main() {
	// Parse command line arguments
	idFlag := flag.Int64("id", 0, "Transcription ID to translate")
	langFlag := flag.String("lang", "ru", "Target language (e.g. 'ru')")
	allFlag := flag.Bool("all", false, "Translate all untranslated transcriptions")
	flag.Parse()

	// Validate arguments
	if *idFlag == 0 && !*allFlag {
		fmt.Println("Must specify either --id or --all")
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
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
	fmt.Printf("Translating transcription ID %d to %s...\n", id, lang)

	err := service.ProcessTranscription(id)
	if err != nil {
		log.Fatalf("Translation error: %v", err)
	}

	fmt.Println("Translation completed and saved to database")
}

// translateAll translates all untranslated transcriptions
func translateAll(lang string, _ *translation.Service) {
	fmt.Printf("Finding untranslated transcriptions for %s translation...\n", lang)

	// TODO: Implement logic to find and translate all untranslated transcriptions
	fmt.Println("Batch translation functionality will be implemented in next version")
}
