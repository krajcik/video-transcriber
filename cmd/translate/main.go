package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"assemblyai-transcriber/internal/config"
	"assemblyai-transcriber/internal/database"
	"assemblyai-transcriber/internal/openrouter"
	"assemblyai-transcriber/internal/translation"
)

func main() {
	// Парсинг аргументов командной строки
	idFlag := flag.Int64("id", 0, "ID транскрипции для перевода")
	langFlag := flag.String("lang", "ru", "Язык перевода (например 'ru')")
	allFlag := flag.Bool("all", false, "Перевести все непереведенные транскрипции")
	flag.Parse()

	// Проверка аргументов
	if *idFlag == 0 && !*allFlag {
		fmt.Println("Необходимо указать либо --id, либо --all")
		flag.Usage()
		os.Exit(1)
	}

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

	// Инициализация клиента OpenRouter
	openrouterClient := openrouter.New(cfg.OpenRouterAPIKey)

	// Создание сервиса перевода
	translationService := translation.New(db, openrouterClient)

	// Выполнение перевода в зависимости от флагов
	if *idFlag > 0 {
		translateSingle(*idFlag, *langFlag, translationService)
	} else if *allFlag {
		translateAll(*langFlag, translationService)
	}
}

// translateSingle переводит одну транскрипцию
func translateSingle(id int64, lang string, service *translation.Service) {
	fmt.Printf("Перевод транскрипции ID %d на %s...\n", id, lang)

	err := service.ProcessTranscription(id)
	if err != nil {
		log.Fatalf("Ошибка перевода: %v", err)
	}

	fmt.Println("Перевод успешно завершен и сохранен в базу данных")
}

// translateAll переводит все непереведенные транскрипции
func translateAll(lang string, service *translation.Service) {
	fmt.Printf("Поиск непереведенных транскрипций для перевода на %s...\n", lang)

	// TODO: Реализовать логику поиска и перевода всех непереведенных транскрипций
	fmt.Println("Функционал перевода всех транскрипций будет реализован в следующей версии")
}
