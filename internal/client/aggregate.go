package client

import (
	"context"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
)

func AggregatePing(client pb.BenchmarkServiceClient, totalRequests int, concurrency int, scenario LoadScenario) {
	log.Println("=== Client Streaming: AggregatePing ===")
	var wg sync.WaitGroup
	wg.Add(concurrency)

	rand.Seed(time.Now().UnixNano())
	var latencies []time.Duration
	var mu sync.Mutex
	var successCount, failCount int64

	for g := 0; g < concurrency; g++ {
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < totalRequests/concurrency; i++ {
				reqStart := time.Now()

				stream, err := client.AggregatePing(context.Background())
				if err != nil {
					mu.Lock()
					failCount++
					mu.Unlock()
					log.Printf("Worker %d: Ошибка AggregatePing: %v", workerID, err)
					continue
				}

				msg := "aggregate ping #" + strconv.Itoa(i+1)
				if err := stream.Send(&pb.PingRequest{Message: msg}); err != nil {
					mu.Lock()
					failCount++
					mu.Unlock()
					log.Printf("Worker %d: Ошибка отправки AggregatePing: %v", workerID, err)
				} else {
					mu.Lock()
					successCount++
					mu.Unlock()
					LogDebug("Worker %d: Отправлено AggregatePing: %s", workerID, msg)
				}

				resp, err := stream.CloseAndRecv()
				if err != nil {
					mu.Lock()
					failCount++
					mu.Unlock()
					log.Printf("Worker %d: Ошибка получения AggregatePing ответа: %v", workerID, err)
				} else {
					latency := time.Since(reqStart)
					mu.Lock()
					latencies = append(latencies, latency)
					mu.Unlock()
					LogDebug("Worker %d: AggregatePing ответ: %s", workerID, resp.Message)
				}

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
		}(g)
	}

	wg.Wait()

	// Расчет p50, p90, p99
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

	log.Printf("AggregatePing: Всего запросов: %d, успешных: %d, неуспешных: %d", totalRequests, successCount, failCount)
	log.Printf("AggregatePing: Latency p50: %s, p90: %s, p99: %s", p50, p90, p99)
}
