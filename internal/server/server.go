package server

import (
	"sync"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
)

// Server структура с метриками
type Server struct {
	pb.UnimplementedBenchmarkServiceServer

	mu        sync.Mutex
	reqCount  int
	totalTime time.Duration
	failCount int
}
