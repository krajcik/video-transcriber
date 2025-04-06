//go:build with_db

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"assemblyai-transcriber/internal/config"
	"assemblyai-transcriber/internal/database"
	"assemblyai-transcriber/internal/openrouter"
	"assemblyai-transcriber/internal/translation"

	assemblyai "github.com/AssemblyAI/assemblyai-go-sdk"
)

func main() {
	// command-line flags
	var (
		translateFlag   = flag.Bool("translate", false, "Enable translation after transcription")
		dbPathFlag      = flag.String("db", "", "Path to SQLite database file")
		openRouterFlag  = flag.String("openrouter-key", "", "OpenRouter API key")
		transcriptFlag  = flag.String("save-transcript", "", "Save transcript to file (specify filename)")
		translationFlag = flag.String("save-translation", "", "Save translation to file (specify filename)")
	)
	flag.Parse()

	// check for required input file argument
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: assemblyai-transcriber [options] <input-file> [api-key]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	inputFile := args[0]
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		log.Fatalf("Input file does not exist: %s", inputFile)
	}

	// load configuration from environment variables and .env
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// override with command-line flags if provided
	if len(args) > 1 {
		cfg.AssemblyAIAPIKey = args[1]
	}
	if *dbPathFlag != "" {
		cfg.DatabasePath = *dbPathFlag
	}
	if *openRouterFlag != "" {
		cfg.OpenRouterAPIKey = *openRouterFlag
	}

	// check for required AssemblyAI API key
	if cfg.AssemblyAIAPIKey == "" {
		log.Fatal("AssemblyAI API key is required. Provide it as second argument or set ASSEMBLYAI_API_KEY in environment or .env file")
	}

	// initialize database
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	if err := db.Setup(); err != nil {
		log.Fatalf("Error setting up database: %v", err)
	}

	// extract audio from input file
	fmt.Println("Extracting audio from video file...")
	audioFile, err := extractAudio(inputFile)
	if err != nil {
		log.Fatalf("Error extracting audio: %v", err)
	}
	defer os.Remove(audioFile) // clean up extracted audio file

	// transcribe audio file
	fmt.Println("Transcribing audio...")
	transcriptText, err := transcribeAudio(audioFile, cfg.AssemblyAIAPIKey)
	if err != nil {
		log.Fatalf("Error transcribing audio: %v", err)
	}

	// save transcription to database
	fmt.Println("Saving transcription to database...")
	transcriptID, err := db.SaveTranscription(filepath.Base(inputFile), transcriptText)
	if err != nil {
		log.Fatalf("Error saving transcription: %v", err)
	}
	fmt.Printf("Transcription saved with ID: %d\n", transcriptID)

	// save transcript to file if requested
	if *transcriptFlag != "" {
		outputPath := *transcriptFlag
		if err := os.WriteFile(outputPath, []byte(transcriptText), 0644); err != nil {
			log.Fatalf("Error saving transcript to file: %v", err)
		}
		fmt.Printf("Transcript saved to: %s\n", outputPath)
	}

	// process translation if requested
	if *translateFlag {
		// check for required OpenRouter API key
		if cfg.OpenRouterAPIKey == "" {
			log.Fatal("OpenRouter API key is required for translation. Set it with --openrouter-key flag or OPENROUTER_API_KEY in environment or .env file")
		}

		// initialize OpenRouter client
		orClient := openrouter.New(cfg.OpenRouterAPIKey)

		// initialize translation service
		translationService := translation.New(db, orClient)

		// process transcription
		fmt.Println("Processing transcription for translation...")
		if err := translationService.ProcessTranscription(transcriptID); err != nil {
			log.Fatalf("Error processing transcription: %v", err)
		}

		// save translation to file if requested
		if *translationFlag != "" {
			if err := translationService.SaveTranslationToFile(transcriptID, *translationFlag); err != nil {
				log.Fatalf("Error saving translation to file: %v", err)
			}
		}
	}

	fmt.Println("Process completed successfully!")
}

// extractAudio extracts audio from a video file
func extractAudio(inputFile string) (string, error) {
	outputFile := "extracted_audio.mp3"
	// add -y flag to automatically overwrite existing files
	cmd := exec.Command("ffmpeg", "-y", "-i", inputFile, "-vn", "-ar", "44.1k", "-ac", "2", "-ab", "128k", "-f", "mp3", outputFile)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ffmpeg error: %v, stderr: %s", err, stderr.String())
	}
	return outputFile, nil
}

// transcribeAudio transcribes an audio file using AssemblyAI
func transcribeAudio(audioFile, apiKey string) (string, error) {
	client := assemblyai.NewClient(apiKey)
	ctx := context.Background()

	// open audio file for reading
	file, err := os.Open(audioFile)
	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// upload file to AssemblyAI server
	fmt.Println("Uploading file to AssemblyAI server...")
	audioURL, err := client.Upload(ctx, file)
	if err != nil {
		return "", fmt.Errorf("error uploading file: %v", err)
	}
	fmt.Println("File uploaded successfully:", audioURL)

	// transcribe audio
	fmt.Println("Starting transcription...")
	transcript, err := client.Transcripts.TranscribeFromURL(ctx, audioURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating transcription: %v", err)
	}

	// wait for transcription completion
	fmt.Println("Waiting for transcription to complete...")
	transcript, err = client.Transcripts.Wait(ctx, assemblyai.ToString(transcript.ID))
	if err != nil {
		return "", fmt.Errorf("error waiting for transcription: %v", err)
	}

	fmt.Println("Transcription completed!")
	return assemblyai.ToString(transcript.Text), nil
}
