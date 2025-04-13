package savetodb

import (
	"context"
	"fmt"
	"path/filepath"

	"assemblyai-transcriber/internal/config"
	"assemblyai-transcriber/internal/interfaces"
)

// ConfigLoader abstracts config loading.
type ConfigLoader func() (*config.Config, error)

// DatabaseFactory abstracts database creation.
type DatabaseFactory func(path string) (interfaces.Database, error)

// TranscriberFactory abstracts transcriber creation.
type TranscriberFactory func(apiKey string) interfaces.Transcriber

// Service provides methods for saving transcripts to the database.
type Service struct {
	ConfigLoader       ConfigLoader
	DatabaseFactory    DatabaseFactory
	TranscriberFactory TranscriberFactory
	FileReader         func(string) ([]byte, error)
}

// NewService creates a new Service instance with dependencies.
func NewService(
	configLoader ConfigLoader,
	dbFactory DatabaseFactory,
	transcriberFactory TranscriberFactory,
	fileReader func(string) ([]byte, error),
) *Service {
	return &Service{
		ConfigLoader:       configLoader,
		DatabaseFactory:    dbFactory,
		TranscriberFactory: transcriberFactory,
		FileReader:         fileReader,
	}
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

	cfg, err := s.ConfigLoader()
	if err != nil {
		return 0, fmt.Errorf("load config: %w", err)
	}

	if opts.DatabasePath != "" {
		cfg.DatabasePath = opts.DatabasePath
	}

	var transcriptText string

	if opts.VideoPath != "" {
		transcriber := s.TranscriberFactory(cfg.AssemblyAIAPIKey)
		transcriptText, err = transcriber.TranscribeVideo(ctx, opts.VideoPath)
		if err != nil {
			return 0, fmt.Errorf("transcribe video: %w", err)
		}
	} else {
		transcriptTextBytes, err := s.FileReader(filepath.Clean(opts.TranscriptPath))
		if err != nil {
			return 0, fmt.Errorf("read transcript file: %w", err)
		}
		transcriptText = string(transcriptTextBytes)
	}

	dbImpl, err := s.DatabaseFactory(cfg.DatabasePath)
	if err != nil {
		return 0, fmt.Errorf("init database: %w", err)
	}
	defer func() {
		_ = dbImpl.Close()
	}()

	if err := dbImpl.Setup(); err != nil {
		return 0, fmt.Errorf("setup database: %w", err)
	}

	var fileName string
	if opts.VideoPath != "" {
		fileName = filepath.Base(opts.VideoPath)
	} else {
		fileName = filepath.Base(opts.TranscriptPath)
	}
	id, err := dbImpl.SaveTranscription(fileName, transcriptText)
	if err != nil {
		return 0, fmt.Errorf("save to database: %w", err)
	}

	return id, nil
}
