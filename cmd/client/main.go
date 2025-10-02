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
	// Устанавливаем соединение с gRPC сервером
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Создаём клиент для BenchmarkService
	client := pb.NewBenchmarkServiceClient(conn)

	const requests = 1000      // общее количество запросов
	const concurrency = 50     // количество параллельных горутин
	var wg sync.WaitGroup
	wg.Add(concurrency)        // ждём завершения всех горутин

	start := time.Now()         // фиксируем начало бенчмарка

	// Запускаем горутины для параллельных запросов
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			// Каждая горутина отправляет часть запросов
			for j := 0; j < requests/concurrency; j++ {
				_, err := client.Ping(context.Background(), &pb.PingRequest{Message: "ping"})
				if err != nil {
					log.Printf("Error: %v", err)
				}
			}
		}()
	}

	wg.Wait()                  // ждём завершения всех горутин
	elapsed := time.Since(start) // вычисляем общее время выполнения

	// Выводим результаты
	log.Printf("Completed %d requests in %s", requests, elapsed)
	log.Printf("RPS: %f", float64(requests)/elapsed.Seconds()) // запросов в секунду
}
