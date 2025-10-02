package main

import (
	"context"
	"log"
	"sync"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewBenchmarkServiceClient(conn)

	const requests = 1000
	const concurrency = 50
	var wg sync.WaitGroup
	wg.Add(concurrency)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < requests/concurrency; j++ {
				_, err := client.Ping(context.Background(), &pb.PingRequest{Message: "ping"})
				if err != nil {
					log.Printf("Error: %v", err)
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)
	log.Printf("Completed %d requests in %s", requests, elapsed)
	log.Printf("RPS: %f", float64(requests)/elapsed.Seconds())
}
