package main

import (
	"context"
	"log"
	"net"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"

	"google.golang.org/grpc"
)

// server реализует gRPC сервис BenchmarkService
type server struct {
	// Встраиваем UnimplementedBenchmarkServiceServer, чтобы реализовать интерфейс
	pb.UnimplementedBenchmarkServiceServer
}

// Ping — простой RPC метод, который возвращает полученное сообщение обратно.
// Это "эхо"-метод, удобный для тестирования и бенчмарка.
func (s *server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	// Возвращаем то же сообщение, что получили
	return &pb.PingResponse{Message: req.Message}, nil
}

func main() {
	// Создаём TCP-листенер на порту 50051
	// Все gRPC соединения будут приниматься на этом порту
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Создаём новый gRPC сервер
	s := grpc.NewServer()

	// Регистрируем наш BenchmarkService на сервере
	pb.RegisterBenchmarkServiceServer(s, &server{})

	log.Println("Server listening on :50051")

	// Запускаем сервер и слушаем входящие соединения
	// Serve блокирует текущую горутину
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
