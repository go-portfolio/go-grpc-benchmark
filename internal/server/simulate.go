package server

import (
	"errors"
	"math/rand"
	"time"
)

// Симуляция обработки запроса: случайная задержка и вероятность ошибки
func SimulateProcessing() (time.Duration, error) {
	delay := time.Duration(1+rand.Intn(5)) * time.Millisecond
	time.Sleep(delay)
	if rand.Float32() < 0.02 { // 2% запросов падают
		return delay, errors.New("simulated server error")
	}
	return delay, nil
}
