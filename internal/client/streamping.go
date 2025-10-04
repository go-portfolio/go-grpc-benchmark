package client

import (
	"context"
	"io"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
)

func StreamPing(client pb.BenchmarkServiceClient, totalRequests int, concurrency int, scenario LoadScenario) {
	log.Println("=== Bidirectional StreamPing ===")
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
				stream, err := client.StreamPing(context.Background())
				if err != nil {
					mu.Lock()
					failCount++
					mu.Unlock()
					log.Printf("Worker %d: Не удалось открыть StreamPing: %v", workerID, err)
					continue
				}

				// Получение ответов
				done := make(chan struct{})
				go func() {
					for {
						resp, err := stream.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							log.Printf("Worker %d: Ошибка получения StreamPing: %v", workerID, err)
							break
						}
						LogDebug("Worker %d: StreamPing response: %s", workerID, resp.Message)
					}
					close(done)
				}()

				msg := "stream ping #" + strconv.Itoa(i+1)
				if err := stream.Send(&pb.PingRequest{Message: msg}); err != nil {
					mu.Lock()
					failCount++
					mu.Unlock()
					log.Printf("Worker %d: Ошибка отправки StreamPing: %v", workerID, err)
				} else {
					mu.Lock()
					successCount++
					mu.Unlock()
					LogDebug("Worker %d: Отправлено StreamPing: %s", workerID, msg)
				}

				if err := stream.CloseSend(); err != nil {
					log.Printf("Worker %d: Ошибка закрытия StreamPing: %v", workerID, err)
				}
				<-done
				latency := time.Since(reqStart)
				mu.Lock()
				latencies = append(latencies, latency)
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

	log.Printf("StreamPing: Всего запросов: %d, успешных: %d, неуспешных: %d", totalRequests, successCount, failCount)
	log.Printf("StreamPing: Latency p50: %s, p90: %s, p99: %s", p50, p90, p99)
}
