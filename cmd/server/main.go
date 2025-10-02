package main

import (
	"flag"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/go-portfolio/go-grpc-benchmark/internal/server"
	"google.golang.org/grpc"
	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
)

var debug bool

func main() {
	flag.BoolVar(&debug, "debug", false, "Enable debug logs")
	flag.Parse()

	log.Println("=== Запуск сервера gRPC Benchmark ===")

	rand.Seed(time.Now().UnixNano())

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Не удалось слушать порт: %v", err)
	}

	s := grpc.NewServer()

	// Передаём флаг debug в сервер
	srv := server.NewServer(debug)
	pb.RegisterBenchmarkServiceServer(s, srv)

	log.Println("Сервер запущен на :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
