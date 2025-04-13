package interfaces

import (
	"context"
)

// Database abstracts database operations.
type Database interface {
	Setup() error
	SaveTranscription(fileName, text string) (int64, error)
	Close() error
}

// Transcriber abstracts transcription operations.
type Transcriber interface {
	TranscribeVideo(ctx context.Context, videoPath string) (string, error)
}
