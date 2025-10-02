package server

import (
	"context"
	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
)

// Unary RPC: Ping
func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	delay, err := SimulateProcessing()
	s.mu.Lock()
	if err != nil {
		s.failCount++
	} else {
		s.reqCount++
		s.totalTime += delay
	}
	s.mu.Unlock()

	if err != nil {
		s.logDebug("Ping error: %v", err)
		return nil, err
	}

	s.logDebug("Ping response: %s", req.Message)
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