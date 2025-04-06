# AssemblyAI Transcriber ![Go Version](https://img.shields.io/github/go-mod/go-version/eshesh/assemblyai-transcriber) ![License](https://img.shields.io/github/license/eshesh/assemblyai-transcriber)

A command-line tool for transcribing audio from video files using AssemblyAI, with additional support for term analysis and Russian translation powered by OpenRouter's Llama 4 Maverick model.

## Table of Contents
- [Features](#features)
- [Requirements](#requirements)  
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Examples](#examples)
- [Translation Process](#translation-process)
  - [Term Management](#term-management-interface)
- [Database Schema](#database-schema)
- [Development](#development)
- [License](#license)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)

## Features

- Extract audio from video files using `ffmpeg`
- Transcribe audio using AssemblyAI API
- Store transcripts in SQLite database
- Analyze text for specialized terms that should not be translated
- Translate transcripts from English to Russian using OpenRouter's Llama 4 Maverick
- Interactive interface for managing untranslatable terms
- Export transcripts and translations to text files

## Requirements

- [Go](https://go.dev/) 1.16 or higher
- [FFmpeg](https://ffmpeg.org/) installed and available in PATH
- [AssemblyAI](https://www.assemblyai.com/) API key
- [OpenRouter](https://openrouter.ai/) API key (for translation features)

## Installation

1. Clone the repository:
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

## Database Schema

```sql
CREATE TABLE transcriptions (
  id TEXT PRIMARY KEY,
  input_file TEXT NOT NULL,
  transcript TEXT NOT NULL, 
  translation TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE terms (
  id INTEGER PRIMARY KEY,
  term TEXT NOT NULL,
  description TEXT,
  transcription_id TEXT REFERENCES transcriptions(id)
);
```

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
