package main

import (
	"context"
	"flag"
	"os"

	"github.com/go-pkgz/lgr"

	"assemblyai-transcriber/internal/config"
	"assemblyai-transcriber/internal/database"
	"assemblyai-transcriber/internal/interfaces"
	"assemblyai-transcriber/internal/savetodb"
	"assemblyai-transcriber/internal/transcribe"
)

func run() int {
	lgr.Setup()
	var (
		transcriptFlag = flag.String("transcript", "", "Path to transcript file")
		videoFlag      = flag.String("video", "", "Path to video file")
		dbPathFlag     = flag.String("db", "", "Path to database file")
	)
	flag.Parse()

	if (*transcriptFlag == "" && *videoFlag == "") || *dbPathFlag == "" {
		lgr.Printf("Usage:")
		lgr.Printf("  For text transcripts: savetodb --transcript=file --db=database.db")
		lgr.Printf("  For video files: savetodb --video=file --db=database.db")
		flag.PrintDefaults()
		return 1
	}

	cfg, err := config.Load()
	if err != nil {
		lgr.Printf("Error loading config: %v", err)
		return 1
	}

	service := savetodb.NewService(
		func() (*config.Config, error) { return cfg, nil },
		func(path string) (interfaces.Database, error) {
			return database.New(path)
		},
		func(apiKey string) interfaces.Transcriber {
			return transcribe.New(apiKey, cfg.MaxAudioFileSizeMB)
		},
		os.ReadFile,
	)
	id, err := service.SaveTranscript(context.Background(), savetodb.SaveTranscriptOptions{
		TranscriptPath: *transcriptFlag,
		VideoPath:      *videoFlag,
		DatabasePath:   *dbPathFlag,
	})
	if err != nil {
		lgr.Printf("Error: %v", err)
		return 1
	}

	lgr.Printf("Successfully saved transcript to database with ID: %d", id)
	return 0
}

func main() {
	os.Exit(run())
}
