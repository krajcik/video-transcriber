# AssemblyAI Transcriber

A command-line tool for transcribing audio from video files using AssemblyAI, with additional support for term analysis and Russian translation powered by OpenRouter's Llama 4 Maverick model.

## Features

- Extract audio from video files using `ffmpeg`
- Transcribe audio using AssemblyAI API
- Store transcripts in SQLite database
- Analyze text for specialized terms that should not be translated
- Translate transcripts from English to Russian using OpenRouter's Llama 4 Maverick
- Interactive interface for managing untranslatable terms
- Export transcripts and translations to text files

## Requirements

- Go 1.16 or higher
- FFmpeg installed and available in PATH
- AssemblyAI API key
- OpenRouter API key (for translation features)

## Installation

1. Clone the repository
2. Install dependencies:

```bash
go get github.com/joho/godotenv
go get github.com/mattn/go-sqlite3
go get github.com/AssemblyAI/assemblyai-go-sdk
```

3. Build the application:

```bash
go build -o assemblyai-transcriber cmd/main.go
```

## Configuration

The application can be configured using:

1. Environment variables
2. `.env` file
3. Command-line flags

Create a `.env` file based on the `.env.example` template:

```
# AssemblyAI API credentials
ASSEMBLYAI_API_KEY=your_api_key_here

# OpenRouter API credentials
OPENROUTER_API_KEY=your_api_key_here

# Database configuration
DATABASE_PATH=./transcriptions.db

# Application settings
LOG_LEVEL=info
```

## Usage

```
./assemblyai-transcriber [options] <input-file> [api-key]

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

### Examples

Basic transcription:
```bash
./assemblyai-transcriber video.mp4
```

Transcription with API key provided as argument:
```bash
./assemblyai-transcriber video.mp4 your_api_key_here
```

Transcription with translation:
```bash
./assemblyai-transcriber -translate video.mp4
```

Save transcript and translation to files:
```bash
./assemblyai-transcriber -translate -save-transcript transcript.txt -save-translation translation.txt video.mp4
```

## Translation Process

When using the translation feature:

1. The application analyzes the transcript to identify specialized terms that should not be translated
2. You are presented with an interactive interface to review and manage these terms
3. The transcript is translated to Russian, preserving the identified terms
4. Both the original transcript and translation are saved to the database

### Term Management Interface

The application provides four options for managing untranslatable terms:

1. Accept all terms as is
2. Reject all terms
3. Process terms interactively (review and manage terms one by one)
4. Edit terms in a text editor

## Database Schema

The application uses SQLite to store:

- Original transcriptions
- Untranslatable terms with descriptions
- Translated texts

## License

This project is licensed under the MIT License - see the LICENSE file for details.
