# AssemblyAI Transcriber

[![Build Status](https://github.com/eshesh/video-transcriber/actions/workflows/ci.yml/badge.svg)](https://github.com/eshesh/video-transcriber/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/eshesh/video-transcriber/branch/master/graph/badge.svg)](https://codecov.io/gh/eshesh/video-transcriber)

A Go application for transcribing video files using AssemblyAI API.

## Features
- Video to text transcription
- Audio extraction from video files
- Translation capabilities

## Installation
1. Clone the repository
2. Set up your AssemblyAI API key in `.env` file
3. Run `go build ./...`

## Usage
```bash
# Transcribe video file to database
./bin/savetodb -input video.mp4

# Translate text
./bin/translate -text "text to translate"
```

## Requirements
- Go 1.24+
- FFmpeg
- AssemblyAI API key

## Development
```bash
# Run all tests
go test ./...

# Run linter
golangci-lint run ./...

# Format code
gofmt -s -w .

# Run with coverage
go test -cover ./...
```
```bash
git clone https://github.com/eshesh/assemblyai-transcriber.git
cd assemblyai-transcriber
```

2. Install dependencies:

```bash
go mod download
```

3. Build the application (options):

```bash
# Main application
go build -o assemblyai-transcriber cmd/main.go

# Or build specific components:
go build -o checkdb cmd/checkdb/main.go
go build -o translate cmd/translate/main.go
```

## Configuration

The application supports multiple configuration methods:

1. **Environment variables** - Highest priority
2. **.env file** - Recommended for local development
3. **Command-line flags** - Override specific settings

Create `.env` file from template:

```ini
# Required
ASSEMBLYAI_API_KEY=your_assemblyai_key
OPENROUTER_API_KEY=your_openrouter_key

# Optional
DATABASE_PATH=./transcriptions.db  # SQLite database path
LOG_LEVEL=info                     # debug/info/warn/error
AUDIO_CACHE_DIR=./audio_cache      # Temporary audio files
```

## Usage [▶️](#examples)

```text
Usage:
  ./assemblyai-transcriber [flags] <input-file>

Flags:
  -db string
        Path to SQLite database file (default "./transcriptions.db")
  -openrouter-key string
        OpenRouter API key
  -save-transcript string
        Save transcript to file
  -save-translation string  
        Save translation to file
  -translate
        Enable translation after transcription
  -v    Enable verbose logging

Options:
  -db string
        Path to SQLite database file
  -openrouter-key string
        OpenRouter API key
  -save-transcript string
        Save transcript to file (specify filename)
  -save-translation string
        Save translation to file (specify filename)
  -translate
        Enable translation after transcription
```

### Examples [↑](#usage)

```bash
# Basic transcription
./assemblyai-transcriber input.mp4

# With translation
./assemblyai-transcriber -translate input.mp4

# Save outputs to files
./assemblyai-transcriber -translate \
  -save-transcript transcript.txt \
  -save-translation translation.txt \
  input.mp4

# Use specific database
./assemblyai-transcriber -db ./custom.db input.mp4
```

## Translation Process

1. **Audio Extraction** - FFmpeg converts video to audio
2. **Transcription** - AssemblyAI processes audio to text
3. **Term Analysis** - Identifies specialized terms
4. **Term Management** - Interactive term review:
   - Accept all terms
   - Reject all terms  
   - Review terms individually
   - Edit terms manually
5. **Translation** - Llama 4 Maverick translates to Russian
6. **Storage** - Results saved to SQLite database

### Term Management Interface

```text
Found 5 specialized terms:
1. API (Application Programming Interface)
2. SQLite (Embedded database)
...

Choose action:
[1] Accept all terms
[2] Reject all terms
[3] Review terms
[4] Edit terms
```

## Database Migrations

Database schema is managed using [goose](https://github.com/pressly/goose) and migration files in the `migrations/` directory.

### Apply migrations

```bash
make migrate-up
```

### Rollback the last migration

```bash
make migrate-down
```

### Check migration status

```bash
make migrate-status
```

By default, migrations are applied to the `data.db` file in the project root (you can change the path in the Makefile).

## Development

Project structure:
```
.
├── cmd/          # CLI commands
│   ├── main.go         # Main app
│   ├── checkdb/        # DB validation
│   └── translate/      # Translation module
├── internal/     # Core packages
│   ├── config/   # Configuration
│   ├── database/ # DB operations  
│   └── ...       # Other components
├── go.mod        # Dependencies
└── go.sum
```

Key packages:
- `internal/config` - Configuration loading
- `internal/database` - SQLite operations
- `internal/openrouter` - Translation API client
- `internal/terms` - Term analysis logic

Build and test:
```bash
# Run tests
go test ./...

# Build with debug symbols
go build -gcflags="all=-N -l" -o assemblyai-transcriber cmd/main.go
```

## License

MIT License - See [LICENSE](LICENSE) for details.

## Troubleshooting

| Issue | Solution |
|-------|----------|
| FFmpeg not found | Install FFmpeg and add to PATH |
| API errors | Verify keys in .env file |
| Database locked | Close other instances using DB |
| Translation fails | Check OpenRouter quota |

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/foo`)
3. Commit your changes (`git commit -am 'Add some foo'`)
4. Push to the branch (`git push origin feature/foo`)  
5. Create a new Pull Request

Please follow the [code conventions](CONVENTIONS.md).
