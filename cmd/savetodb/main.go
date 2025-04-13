package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"assemblyai-transcriber/internal/config"
	"assemblyai-transcriber/internal/database"
	"assemblyai-transcriber/internal/interfaces"
	"assemblyai-transcriber/internal/transcribe"
)

func run() int {
	// command-line flags
	var (
		transcriptFlag = flag.String("transcript", "", "Path to transcript file")
		videoFlag      = flag.String("video", "", "Path to video file")
		dbPathFlag     = flag.String("db", "", "Path to database file")
	)
	flag.Parse()

	// validate required flags
	if (*transcriptFlag == "" && *videoFlag == "") || *dbPathFlag == "" {
		fmt.Println("Usage:")
		fmt.Println("  For text transcripts: savetodb --transcript=file --db=database.db")
		fmt.Println("  For video files: savetodb --video=file --db=database.db")
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

	var transcriptText string

	if *videoFlag != "" {
		// transcribe video file
		var transcriber interfaces.Transcriber = transcribe.New(cfg.AssemblyAIAPIKey)
		transcriptText, err = transcriber.TranscribeVideo(context.Background(), *videoFlag)
		if err != nil {
			log.Printf("Error transcribing video: %v", err)
			return 1
		}
	} else {
		// read transcript file
		transcriptTextBytes, err := os.ReadFile(filepath.Clean(*transcriptFlag))
		if err != nil {
			log.Printf("Error reading transcript file: %v", err)
			return 1
		}
		transcriptText = string(transcriptTextBytes)
	}

	// initialize database
	var db interfaces.Database
	dbImpl, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Printf("Error initializing database: %v", err)
		return 1
	}
	db = dbImpl
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
	fileName := filepath.Base(*videoFlag)
	if fileName == "" {
		fileName = filepath.Base(*transcriptFlag)
	}
	id, err := db.SaveTranscription(fileName, transcriptText)
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
