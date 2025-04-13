package transcribe

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractAudio(t *testing.T) {
	// Mock exec.Command
	execCommand = func(name string, arg ...string) *exec.Cmd {
		assert.Equal(t, "ffmpeg", name)
		assert.Equal(t, []string{
			"-y", "-i", "test.mp4",
			"-vn", "-ar", "44.1k",
			"-ac", "2", "-ab", "128k",
			"-f", "mp3", "extracted_audio.mp3",
		}, arg)
		return exec.Command("echo", "mocked")
	}
	defer func() { execCommand = exec.Command }()

	client := New("test_key", 100)
	_, err := client.extractAudio("test.mp4")
	assert.NoError(t, err)
}
