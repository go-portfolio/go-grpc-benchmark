package client

import (
	"log"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
	"google.golang.org/grpc"
)

func NewBenchmarkClient(target string) pb.BenchmarkServiceClient {
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Не удалось подключиться к серверу: %v", err)
	}
	return pb.NewBenchmarkServiceClient(conn)
}
