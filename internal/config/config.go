package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds all configuration settings
type Config struct {
	AssemblyAIAPIKey string
	OpenRouterAPIKey string
	DatabasePath     string
	LogLevel         string
}

// Load reads the configuration from environment variables
// with optional fallback to .env file
func Load() (*Config, error) {
	// load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		AssemblyAIAPIKey: getEnv("ASSEMBLYAI_API_KEY", ""),
		OpenRouterAPIKey: getEnv("OPENROUTER_API_KEY", ""),
		DatabasePath:     getEnv("DATABASE_PATH", "./transcriptions.db"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
	}

	// ensure database directory exists
	dbDir := filepath.Dir(config.DatabasePath)
	if dbDir != "." {
		if err := os.MkdirAll(dbDir, 0o750); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
