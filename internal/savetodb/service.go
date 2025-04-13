package savetodb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"assemblyai-transcriber/internal/config"
	"assemblyai-transcriber/internal/database"
	"assemblyai-transcriber/internal/interfaces"
	"assemblyai-transcriber/internal/transcribe"
)

 // Service provides methods for saving transcripts to the database.
type Service struct{}

// NewService creates a new Service instance.
func NewService() *Service {
	return &Service{}
}

// SaveTranscriptOptions contains parameters for saving a transcript.
type SaveTranscriptOptions struct {
	TranscriptPath string
	VideoPath      string
	DatabasePath   string
}

 // SaveTranscript saves a transcript (from file or video) to the database.
func (s *Service) SaveTranscript(ctx context.Context, opts SaveTranscriptOptions) (int64, error) {
	if (opts.TranscriptPath == "" && opts.VideoPath == "") || opts.DatabasePath == "" {
		return 0, fmt.Errorf("either transcript or video path and database path must be provided")
	}

	cfg, err := config.Load()
	if err != nil {
		return 0, fmt.Errorf("load config: %w", err)
	}

	if opts.DatabasePath != "" {
		cfg.DatabasePath = opts.DatabasePath
	}

	var transcriptText string

	if opts.VideoPath != "" {
		var transcriber interfaces.Transcriber = transcribe.New(cfg.AssemblyAIAPIKey)
		transcriptText, err = transcriber.TranscribeVideo(ctx, opts.VideoPath)
		if err != nil {
			return 0, fmt.Errorf("transcribe video: %w", err)
		}
	} else {
		transcriptTextBytes, err := os.ReadFile(filepath.Clean(opts.TranscriptPath))
		if err != nil {
			return 0, fmt.Errorf("read transcript file: %w", err)
		}
		transcriptText = string(transcriptTextBytes)
	}

	dbImpl, err := database.New(cfg.DatabasePath)
	if err != nil {
		return 0, fmt.Errorf("init database: %w", err)
	}
	defer func() {
		_ = dbImpl.Close()
	}()

	if err := dbImpl.Setup(); err != nil {
		return 0, fmt.Errorf("setup database: %w", err)
	}

	fileName := filepath.Base(opts.VideoPath)
	if fileName == "" {
		fileName = filepath.Base(opts.TranscriptPath)
	}
	id, err := dbImpl.SaveTranscription(fileName, transcriptText)
	if err != nil {
		return 0, fmt.Errorf("save to database: %w", err)
	}

	return id, nil
}
