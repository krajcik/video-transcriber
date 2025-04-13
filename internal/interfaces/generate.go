//go:build generate
// +build generate

package interfaces

//go:generate moq -pkg=mocks -out=../mocks/database_mock.go . Database
//go:generate moq -pkg=mocks -out=../mocks/transcriber_mock.go . Transcriber
