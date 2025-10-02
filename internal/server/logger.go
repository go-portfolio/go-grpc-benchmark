package server

import (
	"log"
)

var (
	Verbose bool // глобальный флаг для подробного вывода
)

func Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

func Debug(msg string, args ...interface{}) {
	if Verbose {
		log.Printf("[DEBUG] "+msg, args...)
	}
}

func Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}
