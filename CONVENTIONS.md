# Code Conventions and Style Guide

This document outlines the coding conventions and style guidelines used in the AssemblyAI Transcriber project. Following these conventions ensures consistency across the codebase and makes it easier for contributors to understand and maintain the code.

## Directory Structure

The project follows the standard Go project layout:

```
assemblyai-transcriber/
├── cmd/                 # Application entry points
│   └── main.go          # Main application
├── internal/            # Private application code
│   ├── config/          # Configuration handling
│   ├── database/        # Database interactions
│   ├── openrouter/      # OpenRouter API integration
│   ├── terms/           # Term management
│   └── translation/     # Translation functionality
├── pkg/                 # Public libraries that can be used by external applications
├── .env                 # Environment variables (not committed to repository)
├── .env.example         # Example environment variables
├── go.mod               # Go module definition
└── go.sum               # Go module checksums
```

## Naming Conventions

### Files and Directories

- Use lowercase with underscores for file and directory names
- Use descriptive names that reflect the content
- Package names should match their directory names

### Variables and Functions

- Use camelCase for local variables and private functions
- Use PascalCase for exported functions and types
- Choose descriptive names that indicate purpose
- Avoid single-letter variables except for simple loop counters

### Constants

- Use PascalCase for exported constants
- Use all uppercase with underscores for internal constants when representing fixed values

## Code Formatting

- Code should be formatted using `gofmt` or `go fmt`
- Use tabs for indentation (default for Go)
- Maximum line length should be 100 characters when possible
- Group related code blocks with empty lines for readability

## Error Handling

- Always check error returns
- Use the `fmt.Errorf` function with the `%w` verb to wrap errors with context
- Return errors up the call stack with appropriate context
- Avoid using `_` to ignore errors except in very specific cases

Example:
```go
if err := someFunction(); err != nil {
    return fmt.Errorf("failed to execute someFunction: %w", err)
}
```

## Comments and Documentation

- Every exported function should have a comment in the format that will render correctly in godoc
- Start comments with the name of the thing being described
- Write comments in complete sentences with proper capitalization and punctuation
- Add inline comments for complex or non-obvious code sections
- Use // for single-line comments and /* */ for multi-line comments

Example:
```go
// SaveTranscription saves a transcription to the database and returns the ID.
// It returns an error if the database operation fails.
func SaveTranscription(fileName, text string) (int64, error) {
    // Implementation details here
}
```

## Imports

- Group imports into standard library, external packages, and internal packages
- Sort imports alphabetically within each group
- Remove unused imports

Example:
```go
import (
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/joho/godotenv"
    "github.com/mattn/go-sqlite3"
    
    "assemblyai-transcriber/internal/config"
)
```

## Testing

- Write tests for all exported functions
- Use table-driven tests where appropriate
- Name test files with the `_test.go` suffix
- Use the `testing` package for unit tests

## Dependency Management

- Use Go modules for dependency management
- Explicitly specify versions in go.mod
- Regularly update dependencies to incorporate security fixes

## Configuration

- Use environment variables for configuration
- Provide sensible defaults when possible
- Support .env files for local development
- Validate configuration at startup

## Error Messages

- Error messages should be clear, specific, and actionable
- Include relevant details that would help troubleshoot the issue
- Keep error messages consistent in style and format

## Logging

- Use the standard `log` package or a compatible logging library
- Include relevant context in log messages
- Use appropriate log levels (info, warning, error)
- Be consistent with log message formatting

## Version Control

- Write clear, concise commit messages
- Use the imperative mood in commit messages (e.g., "Add feature" not "Added feature")
- Keep commits focused on a single change or feature
- Reference issue numbers in commit messages when applicable

By following these conventions, we maintain a consistent, readable, and maintainable codebase that's easy for all contributors to work with.
