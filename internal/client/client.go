package client

import (
    pb "github.com/go-portfolio/go-grpc-benchmark/proto"
    "google.golang.org/grpc"
)

type BenchmarkClient struct {
    pb.BenchmarkServiceClient
}

func NewBenchmarkClientWithConn(conn *grpc.ClientConn) *BenchmarkClient {
    return &BenchmarkClient{
        BenchmarkServiceClient: pb.NewBenchmarkServiceClient(conn),
    }
}
