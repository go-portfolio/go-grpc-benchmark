package client

import (
	"io"
	"log"
	"os"
	"sync"
)

// Глобальные флаги
var (
	Debug   bool
	Verbose bool
)

var (
	logger   *log.Logger
	logFile  *os.File
	initOnce sync.Once
)

// InitLogger инициализирует логирование
// filePath — путь к файлу логов (например "./logs/client.log")
func InitLogger(filePath string) error {
	var err error
	initOnce.Do(func() {
		// Создаём папку, если её нет
		if err = os.MkdirAll("./logs", 0755); err != nil {
			return
		}

		// Открываем файл логов (добавляем записи, не перезаписываем)
		logFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}

		// Настраиваем логгер: запись и в файл, и в stdout
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		logger = log.New(multiWriter, "", log.LstdFlags|log.Lmicroseconds)
	})
	return err
}

// Закрываем файл логов при завершении программы
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// LogDebug пишет лог только при включённом режиме Debug
func LogDebug(format string, v ...interface{}) {
	if Debug && logger != nil {
		logger.Printf("[DEBUG] "+format, v...)
	}
}

// LogVerbose пишет лог только при включённом режиме Verbose
func LogVerbose(format string, v ...interface{}) {
	if Verbose && logger != nil {
		logger.Printf("[VERBOSE] "+format, v...)
	}
}

// LogInfo — просто инфо-логи (всегда выводятся)
func LogInfo(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[INFO] "+format, v...)
	}
}
