package client

import (
	"context"
	"log"
	"sync"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
)

func UnaryPing(client pb.BenchmarkServiceClient, requests int, concurrency int) {
	log.Println("=== Бенчмарк Unary Ping ===")
	var wg sync.WaitGroup
	wg.Add(concurrency)

	var successCount int64
	var failCount int64
	var mu sync.Mutex

	start := time.Now()
	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < requests/concurrency; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				resp, err := client.Ping(ctx, &pb.PingRequest{Message: "ping"})
				cancel()

				mu.Lock()
				if err != nil {
					failCount++
					LogDebug("Worker %d: Ping error: %v", workerID, err)
					LogVerbose("Ping failed: %v", err)
				} else {
					successCount++
					LogDebug("Worker %d: Ping response: %s", workerID, resp.Message)
					LogVerbose("Ping succeeded")
				}
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	log.Printf("Всего запросов: %d, успешных: %d, неуспешных: %d", requests, successCount, failCount)
	log.Printf("Общее время выполнения: %s", elapsed)
	log.Printf("Средняя скорость (RPS): %.2f", float64(successCount)/elapsed.Seconds())
}
