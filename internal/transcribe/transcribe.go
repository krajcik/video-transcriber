package transcribe

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	assemblyai "github.com/AssemblyAI/assemblyai-go-sdk"
)

var execCommand = exec.Command

// Client is a client for AssemblyAI API
type Client struct {
	apiKey string
}

// New creates a new transcription client
func New(apiKey string) *Client {
	return &Client{apiKey: apiKey}
}

// TranscribeVideo performs video file transcription
func (c *Client) TranscribeVideo(ctx context.Context, videoPath string) (string, error) {
	// Extract audio from video
	audioPath, err := c.extractAudio(videoPath)
	if err != nil {
		return "", fmt.Errorf("audio extraction error: %v", err)
	}
	defer os.Remove(audioPath)

	// Transcribe audio
	return c.transcribeAudio(ctx, audioPath)
}

// extractAudio extracts audio from video file using ffmpeg
func (c *Client) extractAudio(videoPath string) (string, error) {
	audioPath := "extracted_audio.mp3"
	cmd := execCommand("ffmpeg", "-y", "-i", videoPath, "-vn", "-ar", "44.1k", "-ac", "2", "-ab", "128k", "-f", "mp3", audioPath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg error: %v, stderr: %s", err, stderr.String())
	}
	return audioPath, nil
}

// transcribeAudio performs transcription using AssemblyAI API
func (c *Client) transcribeAudio(ctx context.Context, audioPath string) (string, error) {
	client := assemblyai.NewClient(c.apiKey)

	file, err := os.Open(filepath.Clean(audioPath))
	if err != nil {
		return "", fmt.Errorf("file open error: %v", err)
	}
	defer file.Close()

	// Upload file to AssemblyAI server
	audioURL, err := client.Upload(ctx, file)
	if err != nil {
		return "", fmt.Errorf("file upload error: %v", err)
	}

	// Start transcription
	transcript, err := client.Transcripts.TranscribeFromURL(ctx, audioURL, nil)
	if err != nil {
		return "", fmt.Errorf("transcription start error: %v", err)
	}

	// Wait for transcription completion
	transcript, err = client.Transcripts.Wait(ctx, assemblyai.ToString(transcript.ID))
	if err != nil {
		return "", fmt.Errorf("transcription wait error: %v", err)
	}

	return assemblyai.ToString(transcript.Text), nil
}
