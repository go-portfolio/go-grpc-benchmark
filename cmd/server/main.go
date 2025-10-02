package main

import (
	"flag"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/go-portfolio/go-grpc-benchmark/internal/server"
	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
	"google.golang.org/grpc"
)

func main() {
	// ---------------- Флаги ----------------
	debug := flag.Bool("debug", false, "Enable debug mode")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	if *debug {
		log.Println("[INFO] Debug mode enabled")
	}
	if *verbose {
		log.Println("[INFO] Verbose logging enabled")
	}

	rand.Seed(time.Now().UnixNano())

	// ---------------- Создание сервера ----------------
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Не удалось слушать порт: %v", err)
	}

	// Создаем сервер с передачей флагов debug и verbose
	srv := server.NewServer(*debug, *verbose)

	grpcServer := grpc.NewServer()
	pb.RegisterBenchmarkServiceServer(grpcServer, srv)

	log.Println("[INFO] Сервер запущен на :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
