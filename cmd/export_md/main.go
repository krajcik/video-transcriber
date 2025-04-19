package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-pkgz/lgr"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"assemblyai-transcriber/internal/config"
)

func main() {
	lgr.Setup()
	outDir := flag.String("out", "./translations", "Output directory for MD files")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		lgr.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize database
	db, err := sqlx.Open("sqlite3", cfg.DatabasePath)
	if err != nil {
		lgr.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// Create output directory
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		lgr.Fatalf("Error creating output directory: %v", err)
	}

	// Get all translations
	var translations []struct {
		ID             int    `db:"id"`
		TranslatedText string `db:"translated_text"`
	}

	err = db.Select(&translations, "SELECT id, translated_text FROM translations")
	if err != nil {
		lgr.Fatalf("Error querying translations: %v", err)
	}

	lgr.Printf("Found %d translations", len(translations))

	// Save each translation to a markdown file
	for _, t := range translations {
		outputPath := filepath.Join(*outDir, fmt.Sprintf("translation_%d.md", t.ID))
		err = os.WriteFile(outputPath, []byte(t.TranslatedText), 0644)
		if err != nil {
			lgr.Printf("Error saving %s: %v", outputPath, err)
			continue
		}
		lgr.Printf("Saved %s", outputPath)
	}

	lgr.Printf("Done!")
}