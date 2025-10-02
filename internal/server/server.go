package server

import (
	"log"
	"sync"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
)

type Server struct {
	pb.UnimplementedBenchmarkServiceServer

	mu        sync.Mutex
	verbose   bool
	debug     bool
	reqCount  int
	totalTime time.Duration
	failCount int
}

// Конструктор сервера с debug и verbose флагами
func NewServer(debug, verbose bool) *Server {
	return &Server{
		debug:   debug,
		verbose: verbose,
	}
}

// Вспомогательная функция для вывода debug-логов
func (s *Server) logDebug(format string, v ...interface{}) {
	if s.debug {
		log.Printf("[DEBUG] "+format, v...)
	}
}

// Вспомогательная функция для вывода verbose-логов
func (s *Server) logVerbose(format string, v ...interface{}) {
	if s.verbose {
		log.Printf("[VERBOSE] "+format, v...)
	}
}
