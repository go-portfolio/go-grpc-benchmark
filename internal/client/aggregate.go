package client

import (
	"context"
	"log"
	"strconv"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
)

func AggregatePing(client pb.BenchmarkServiceClient, count int) {
	log.Println("=== Client Streaming: AggregatePing ===")
	stream, err := client.AggregatePing(context.Background())
	if err != nil {
		log.Fatalf("Ошибка AggregatePing: %v", err)
	}

	for i := 1; i <= count; i++ {
		msg := "aggregate ping #" + strconv.Itoa(i)
		if err := stream.Send(&pb.PingRequest{Message: msg}); err != nil {
			log.Printf("Ошибка отправки AggregatePing: %v", err)
		} else {
			LogDebug("Отправлено AggregatePing: %s", msg)
		}
		time.Sleep(50 * time.Millisecond)
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Printf("Ошибка получения AggregatePing ответа: %v", err)
	} else {
		LogDebug("AggregatePing ответ: %s", resp.Message)
	}
}
