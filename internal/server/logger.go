package server

import (
	"io"
	"log"
	"os"
	"sync"
)

// Глобальные флаги
var (
	IsDebug bool // режим отладки
	Verbose bool // режим подробных логов
)

// Локальные переменные для логгера
var (
	logger   *log.Logger
	logFile  *os.File
	initOnce sync.Once
)

// InitLogger инициализирует логирование.
// filePath — путь к файлу логов (например "./logs/server.log")
func InitLogger(filePath string) error {
	var err error
	initOnce.Do(func() {
		// Создаём каталог logs, если он отсутствует
		if err = os.MkdirAll("./logs", 0755); err != nil {
			return
		}

		// Открываем файл логов (в режиме дозаписи)
		logFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}

		// Настраиваем логгер: вывод одновременно в консоль и в файл
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		logger = log.New(multiWriter, "", log.LstdFlags|log.Lmicroseconds)
	})
	return err
}

// CloseLogger — корректное закрытие файла при завершении
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// Info — обычное информационное сообщение (всегда выводится)
func Info(msg string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[INFO] "+msg, args...)
	}
}

// Debug — сообщение отладки, выводится только при включённом IsDebug или Verbose
func Debug(msg string, args ...interface{}) {
	if (IsDebug || Verbose) && logger != nil {
		logger.Printf("[DEBUG] "+msg, args...)
	}
}

// Error — сообщение об ошибке (всегда выводится)
func Error(msg string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[ERROR] "+msg, args...)
	}
}
