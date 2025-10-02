package main

import (
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/go-portfolio/go-grpc-benchmark/internal/server"
	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
	"google.golang.org/grpc"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Не удалось слушать порт: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterBenchmarkServiceServer(s, &server.Server{})

	log.Println("Сервер запущен на :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
