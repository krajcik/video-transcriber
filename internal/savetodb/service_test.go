package savetodb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"assemblyai-transcriber/internal/config"
	"assemblyai-transcriber/internal/interfaces"
	"assemblyai-transcriber/internal/mocks"
)

func TestService_SaveTranscript_File_Success(t *testing.T) {
	mockDB := &mocks.DatabaseMock{
		SetupFunc: func() error { return nil },
		SaveTranscriptionFunc: func(fileName, text string) (int64, error) {
			require.Equal(t, "transcript.txt", fileName)
			require.Equal(t, "test transcript", text)
			return 42, nil
		},
		CloseFunc: func() error { return nil },
	}
	service := NewService(
		func() (*config.Config, error) {
			return &config.Config{DatabasePath: "test.db"}, nil
		},
		func(path string) (interfaces.Database, error) {
			require.Equal(t, "test.db", path)
			return mockDB, nil
		},
		nil, // TranscriberFactory not used
		func(path string) ([]byte, error) {
			require.Equal(t, "transcript.txt", path)
			return []byte("test transcript"), nil
		},
	)

	id, err := service.SaveTranscript(context.Background(), SaveTranscriptOptions{
		TranscriptPath: "transcript.txt",
		DatabasePath:   "test.db",
	})
	require.NoError(t, err)
	require.Equal(t, int64(42), id)
}

func TestService_SaveTranscript_Video_Success(t *testing.T) {
	mockTranscriber := &mocks.TranscriberMock{
		TranscribeVideoFunc: func(ctx context.Context, videoPath string) (string, error) {
			require.Equal(t, "video.mp4", videoPath)
			return "video transcript", nil
		},
	}
	mockDB := &mocks.DatabaseMock{
		SetupFunc: func() error { return nil },
		SaveTranscriptionFunc: func(fileName, text string) (int64, error) {
			require.Equal(t, "video.mp4", fileName)
			require.Equal(t, "video transcript", text)
			return 99, nil
		},
		CloseFunc: func() error { return nil },
	}
	service := NewService(
		func() (*config.Config, error) {
			return &config.Config{DatabasePath: "test.db", AssemblyAIAPIKey: "key"}, nil
		},
		func(path string) (interfaces.Database, error) {
			require.Equal(t, "test.db", path)
			return mockDB, nil
		},
		func(apiKey string) interfaces.Transcriber {
			require.Equal(t, "key", apiKey)
			return mockTranscriber
		},
		nil, // FileReader not used
	)

	id, err := service.SaveTranscript(context.Background(), SaveTranscriptOptions{
		VideoPath:    "video.mp4",
		DatabasePath: "test.db",
	})
	require.NoError(t, err)
	require.Equal(t, int64(99), id)
}
