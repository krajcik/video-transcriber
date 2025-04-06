package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_Load(t *testing.T) {
	tests := []struct {
		name     string
		envSetup func()
		cleanup  func()
		want     *Config
		wantErr  bool
	}{
		{
			name: "default values",
			envSetup: func() {
				os.Clearenv()
			},
			want: &Config{
				AssemblyAIAPIKey: "",
				OpenRouterAPIKey: "",
				DatabasePath:     "./transcriptions.db",
				LogLevel:         "info",
			},
		},
		{
			name: "from environment",
			envSetup: func() {
				os.Clearenv()
				os.Setenv("ASSEMBLYAI_API_KEY", "test_assemblyai")
				os.Setenv("OPENROUTER_API_KEY", "test_openrouter")
				os.Setenv("DATABASE_PATH", "/tmp/test.db")
				os.Setenv("LOG_LEVEL", "debug")
			},
			want: &Config{
				AssemblyAIAPIKey: "test_assemblyai",
				OpenRouterAPIKey: "test_openrouter",
				DatabasePath:     "/tmp/test.db",
				LogLevel:         "debug",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.envSetup != nil {
				tt.envSetup()
			}

			// Test
			got, err := Load()

			// Verify
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)

			// Cleanup
			if tt.cleanup != nil {
				tt.cleanup()
			}
		})
	}
}

func TestConfig_LoadWithEnvFile(t *testing.T) {
	// Clear env before test
	os.Clearenv()

	// Create temp .env file in current directory
	envPath := ".env"
	err := os.WriteFile(envPath, []byte(`
ASSEMBLYAI_API_KEY=env_test_assemblyai
OPENROUTER_API_KEY=env_test_openrouter
DATABASE_PATH=env_test.db
LOG_LEVEL=warn
`), 0o600)
	require.NoError(t, err)
	defer os.Remove(envPath)

	// Test
	got, err := Load()
	require.NoError(t, err)

	// Only check fields we explicitly set in .env
	require.Equal(t, "env_test_assemblyai", got.AssemblyAIAPIKey)
	require.Equal(t, "env_test_openrouter", got.OpenRouterAPIKey)
	require.Equal(t, "warn", got.LogLevel)
}

func Test_getEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		setup    func()
		defValue string
		want     string
	}{
		{
			name:     "env not set",
			key:      "TEST_KEY",
			defValue: "default",
			want:     "default",
		},
		{
			name: "env set",
			key:  "TEST_KEY",
			setup: func() {
				os.Setenv("TEST_KEY", "test_value")
			},
			defValue: "default",
			want:     "test_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			if tt.setup != nil {
				tt.setup()
			}
			got := getEnv(tt.key, tt.defValue)
			require.Equal(t, tt.want, got)
		})
	}
}
