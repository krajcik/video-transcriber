package main

import (
	"fmt"
	"log"

	"assemblyai-transcriber/internal/config"
	"assemblyai-transcriber/internal/database"
)

func main() {
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

	// Check transcription with ID=1
	_, err = db.GetTranscription(1)
	if err != nil {
		log.Printf("Transcription with ID=1 not found: %v", err)
	} else {
		fmt.Println("Transcription with ID=1 exists in database")
	}
}
