package client

import (
	"context"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
)

func UnaryPing(client pb.BenchmarkServiceClient, requests int, concurrency int, scenario LoadScenario) {
	log.Println("=== Бенчмарк Unary Ping ===")
	var wg sync.WaitGroup
	wg.Add(concurrency)

	var successCount int64
	var failCount int64
	var mu sync.Mutex
	var latencies []time.Duration

	start := time.Now()
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < requests/concurrency; j++ {
				reqStart := time.Now()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				resp, err := client.Ping(ctx, &pb.PingRequest{Message: "ping"})
				cancel()
				reqLatency := time.Since(reqStart)

				mu.Lock()
				if err != nil {
					failCount++
					LogDebug("Worker %d: Ping error: %v", workerID, err)
					LogVerbose("Ping failed: %v", err)
				} else {
					successCount++
					latencies = append(latencies, reqLatency)
					LogDebug("Worker %d: Ping response: %s", workerID, resp.Message)
					LogVerbose("Ping succeeded")
				}
				mu.Unlock()

				// Симуляция сценариев нагрузки
				switch scenario {
				case ScenarioLight:
					time.Sleep(100 * time.Millisecond)
				case ScenarioPeak:
					time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
				case ScenarioConstant:
					// без пауз
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Расчёт p50, p90, p99
	var p50, p90, p99 time.Duration
	if len(latencies) > 0 {
		sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
		idx := func(p float64) int {
			i := int(float64(len(latencies)) * p / 100.0)
			if i >= len(latencies) {
				i = len(latencies) - 1
			}
			return i
		}
		p50 = latencies[idx(50)]
		p90 = latencies[idx(90)]
		p99 = latencies[idx(99)]
	}

	log.Printf("Всего запросов: %d, успешных: %d, неуспешных: %d", requests, successCount, failCount)
	log.Printf("Общее время выполнения: %s", elapsed)
	log.Printf("Средняя скорость (RPS): %.2f", float64(successCount)/elapsed.Seconds())
	log.Printf("Latency p50: %s, p90: %s, p99: %s", p50, p90, p99)
}
