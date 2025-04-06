package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"assemblyai-transcriber/internal/config"
	"assemblyai-transcriber/internal/database"
)

func run() int {
	// command-line flags
	var (
		transcriptFlag = flag.String("transcript", "", "Path to transcript file")
		dbPathFlag     = flag.String("db", "", "Path to database file")
	)
	flag.Parse()

	// validate required flags
	if *transcriptFlag == "" || *dbPathFlag == "" {
		fmt.Println("Usage: savetodb --transcript=file --db=database.db")
		flag.PrintDefaults()
		return 1
	}

	// load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Error loading configuration: %v", err)
		return 1
	}

	// override db path if specified
	if *dbPathFlag != "" {
		cfg.DatabasePath = *dbPathFlag
	}

	// read transcript file
	transcriptText, err := os.ReadFile(filepath.Clean(*transcriptFlag))
	if err != nil {
		log.Printf("Error reading transcript file: %v", err)
		return 1
	}

	// initialize database
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Printf("Error initializing database: %v", err)
		return 1
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// setup database schema
	if err := db.Setup(); err != nil {
		log.Printf("Error setting up database: %v", err)
		return 1
	}

	// save to database
	fileName := filepath.Base(*transcriptFlag)
	id, err := db.SaveTranscription(fileName, string(transcriptText))
	if err != nil {
		log.Printf("Error saving to database: %v", err)
		return 1
	}

	fmt.Printf("Successfully saved transcript to database with ID: %d\n", id)
	return 0
}

func main() {
	os.Exit(run())
}
