package client

import (
    pb "github.com/go-portfolio/go-grpc-benchmark/proto"
    "google.golang.org/grpc"
)

type BenchmarkClient struct {
    pb.BenchmarkServiceClient
}

type LoadScenario string

const (
	ScenarioLight    LoadScenario = "light"
	ScenarioPeak     LoadScenario = "peak"
	ScenarioConstant LoadScenario = "constant"
)

func NewBenchmarkClientWithConn(conn *grpc.ClientConn) *BenchmarkClient {
    return &BenchmarkClient{
        BenchmarkServiceClient: pb.NewBenchmarkServiceClient(conn),
    }
}
