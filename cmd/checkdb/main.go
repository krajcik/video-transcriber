package main

import (
	"fmt"
	"log"

	"assemblyai-transcriber/internal/config"
	"assemblyai-transcriber/internal/database"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализация базы данных
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer db.Close()

	// Проверка транскрипции с ID=1
	_, err = db.GetTranscription(1)
	if err != nil {
		log.Printf("Транскрипция с ID=1 не найдена: %v", err)
	} else {
		fmt.Println("Транскрипция с ID=1 существует в базе данных")
	}
}
