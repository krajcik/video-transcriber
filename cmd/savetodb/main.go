package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"assemblyai-transcriber/internal/config"
	"assemblyai-transcriber/internal/database"
)

func main() {
	// Загрузка конфигурации из переменных среды и .env
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Проверяем аргументы командной строки
	if len(os.Args) < 2 {
		fmt.Println("Использование: savetodb <путь-к-транскрипции> [путь-к-базе-данных]")
		os.Exit(1)
	}

	// Путь к файлу транскрипции
	transcriptPath := os.Args[1]

	// Проверяем существование файла
	if _, err := os.Stat(transcriptPath); os.IsNotExist(err) {
		log.Fatalf("Файл транскрипции не существует: %s", transcriptPath)
	}

	// Читаем содержимое файла
	transcriptText, err := os.ReadFile(transcriptPath)
	if err != nil {
		log.Fatalf("Ошибка чтения файла транскрипции: %v", err)
	}

	// Используем путь к базе данных из аргументов или из конфигурации
	dbPath := cfg.DatabasePath
	if len(os.Args) > 2 {
		dbPath = os.Args[2]
	}

	// Инициализация базы данных
	fmt.Println("Подключение к базе данных:", dbPath)
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer db.Close()

	// Настройка схемы базы данных
	if err := db.Setup(); err != nil {
		log.Fatalf("Ошибка настройки базы данных: %v", err)
	}

	// Сохранение транскрипции в базу данных
	fmt.Println("Сохранение транскрипции в базу данных...")
	fileName := filepath.Base(transcriptPath)
	transcriptID, err := db.SaveTranscription(fileName, string(transcriptText))
	if err != nil {
		log.Fatalf("Ошибка сохранения транскрипции: %v", err)
	}

	fmt.Printf("Транскрипция успешно сохранена с ID: %d\n", transcriptID)
}
