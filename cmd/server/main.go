package main

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"

	"google.golang.org/grpc"
)

// server реализует gRPC сервис BenchmarkService
type server struct {
	pb.UnimplementedBenchmarkServiceServer

	mu        sync.Mutex
	reqCount  int
	totalTime time.Duration
}

// Ping — простой RPC метод, который возвращает полученное сообщение обратно.
func (s *server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	start := time.Now()

	resp := &pb.PingResponse{Message: req.Message}

	// обновляем статистику
	elapsed := time.Since(start)
	s.mu.Lock()
	s.reqCount++
	s.totalTime += elapsed
	s.mu.Unlock()

	return resp, nil
}

// Stats — метод, возвращающий статистику по вызовам
func (s *server) Stats(ctx context.Context, req *pb.StatsRequest) (*pb.StatsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var avg float64
	if s.reqCount > 0 {
		avg = s.totalTime.Seconds() / float64(s.reqCount)
	}

	return &pb.StatsResponse{
		TotalRequests: int32(s.reqCount),
		AvgLatencySec: avg,
	}, nil
}

// StreamPing — двунаправленный стриминг (клиент ↔ сервер)
func (s *server) StreamPing(stream pb.BenchmarkService_StreamPingServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		// сразу же отвечаем
		if err := stream.Send(&pb.PingResponse{Message: "echo: " + req.Message}); err != nil {
			return err
		}
	}
}

func main() {
	// Создаём TCP-листенер на порту 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Создаём новый gRPC сервер
	s := grpc.NewServer()

	// Регистрируем BenchmarkService
	pb.RegisterBenchmarkServiceServer(s, &server{})

	log.Println("Server listening on :50051")

	// Запускаем сервер (блокирует текущую горутину)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
