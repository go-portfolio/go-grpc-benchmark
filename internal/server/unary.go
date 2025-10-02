package server

import (
	"context"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
)

// Пример использования внутри метода Ping
func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	start := time.Now()
	s.logVerbose("Received Ping: %s", req.Message)

	// Имитация обработки (можно добавить simulateProcessing и т.д.)
	time.Sleep(5 * time.Millisecond)

	elapsed := time.Since(start)
	s.mu.Lock()
	s.reqCount++
	s.totalTime += elapsed
	s.mu.Unlock()

	s.logDebug("Ping processed in %v", elapsed)
	return &pb.PingResponse{Message: req.Message}, nil
}

// Unary RPC: Stats
func (s *Server) Stats(ctx context.Context, req *pb.StatsRequest) (*pb.StatsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	avg := 0.0
	if s.reqCount > 0 {
		avg = s.totalTime.Seconds() / float64(s.reqCount)
	}
	s.logDebug("Stats requested: total=%d avg=%.6f", s.reqCount, avg)
	return &pb.StatsResponse{
		TotalRequests: int32(s.reqCount),
		AvgLatencySec: avg,
	}, nil
}
