package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"assemblyai-transcriber/internal/config"
	"assemblyai-transcriber/internal/database"
	"assemblyai-transcriber/internal/interfaces"
	"assemblyai-transcriber/internal/savetodb"
	"assemblyai-transcriber/internal/transcribe"
)

func run() int {
	var (
		transcriptFlag = flag.String("transcript", "", "Path to transcript file")
		videoFlag      = flag.String("video", "", "Path to video file")
		dbPathFlag     = flag.String("db", "", "Path to database file")
	)
	flag.Parse()

	if (*transcriptFlag == "" && *videoFlag == "") || *dbPathFlag == "" {
		fmt.Println("Usage:")
		fmt.Println("  For text transcripts: savetodb --transcript=file --db=database.db")
		fmt.Println("  For video files: savetodb --video=file --db=database.db")
		flag.PrintDefaults()
		return 1
	}

	service := savetodb.NewService(
		config.Load,
		func(path string) (interfaces.Database, error) {
			return database.New(path)
		},
		func(apiKey string) interfaces.Transcriber {
			return transcribe.New(apiKey)
		},
		os.ReadFile,
	)
	id, err := service.SaveTranscript(context.Background(), savetodb.SaveTranscriptOptions{
		TranscriptPath: *transcriptFlag,
		VideoPath:      *videoFlag,
		DatabasePath:   *dbPathFlag,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Printf("Successfully saved transcript to database with ID: %d\n", id)
	return 0
}

func main() {
	os.Exit(run())
}
