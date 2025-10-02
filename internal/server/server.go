package server

import (
	"context"
	"io"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
)

type Server struct {
	pb.UnimplementedBenchmarkServiceServer

	mu        sync.Mutex
	reqCount  int
	totalTime time.Duration
	failCount int
	debug     bool
}

// Конструктор сервера с debug-флагом
func NewServer(debug bool) *Server {
	return &Server{debug: debug}
}

// Вспомогательная функция для вывода debug-логов
func (s *Server) logDebug(format string, v ...interface{}) {
	if s.debug {
		log.Printf("[DEBUG] "+format, v...)
	}
}
