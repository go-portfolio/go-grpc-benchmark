package main

import (
	"context"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
	"google.golang.org/grpc"
)

func main() {
	// ---------------- Подключение к серверу ----------------
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure()) // локальное подключение без TLS
	if err != nil {
		log.Fatalf("Не удалось подключиться к серверу: %v", err)
	}
	defer conn.Close()

	client := pb.NewBenchmarkServiceClient(conn)

	// ================= Unary RPC: Ping =================
	const requests = 1000   // общее количество запросов
	const concurrency = 50  // количество параллельных горутин

	log.Printf("=== Бенчмарк Unary Ping ===")
	var wg sync.WaitGroup
	wg.Add(concurrency)
	start := time.Now()

	var successCount int64
	var failCount int64
	var mu sync.Mutex

	// Параллельные горутины для отправки Ping
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < requests/concurrency; j++ {
				_, err := client.Ping(context.Background(), &pb.PingRequest{Message: "ping"})
				mu.Lock()
				if err != nil {
					failCount++
					log.Printf("Ошибка Ping: %v", err)
				} else {
					successCount++
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	log.Printf("Бенчмарк Ping завершён")
	log.Printf("Всего запросов: %d, успешных: %d, неуспешных: %d", requests, successCount, failCount)
	log.Printf("Общее время выполнения: %s", elapsed)
	log.Printf("Средняя скорость (RPS): %.2f запросов/сек", float64(successCount)/elapsed.Seconds())

	// ================= Unary RPC: Stats =================
	stats, err := client.Stats(context.Background(), &pb.StatsRequest{})
	if err != nil {
		log.Fatalf("Не удалось получить статистику сервера: %v", err)
	}
	log.Printf("Статистика сервера: всего обработано %d запросов, средняя задержка %.6f сек",
		stats.TotalRequests, stats.AvgLatencySec)

	// ================= Bidirectional Streaming: StreamPing =================
	log.Println("=== Bidirectional StreamPing ===")
	stream, err := client.StreamPing(context.Background())
	if err != nil {
		log.Fatalf("Не удалось открыть StreamPing: %v", err)
	}

	// Горутинa для приёма сообщений от сервера
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				log.Println("StreamPing: поток завершён сервером")
				return
			}
			if err != nil {
				log.Printf("Ошибка получения StreamPing: %v", err)
				return
			}
			log.Printf("Ответ StreamPing: %s", resp.Message)
		}
	}()

	// Отправка сообщений через bidirectional stream
	for i := 1; i <= 5; i++ {
		msg := "stream ping #" + strconv.Itoa(i)
		if err := stream.Send(&pb.PingRequest{Message: msg}); err != nil {
			log.Printf("Ошибка отправки StreamPing: %v", err)
		} else {
			log.Printf("Отправлено StreamPing: %s", msg)
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err := stream.CloseSend(); err != nil {
		log.Printf("Ошибка закрытия отправки StreamPing: %v", err)
	}
	time.Sleep(500 * time.Millisecond) // ждём получение всех ответов

	// ================= Server Streaming: PushNotifications =================
	log.Println("=== Server Streaming: PushNotifications ===")
	notifyStream, err := client.PushNotifications(context.Background(), &pb.PingRequest{Message: "start"})
	if err != nil {
		log.Fatalf("Ошибка PushNotifications: %v", err)
	}
	for {
		resp, err := notifyStream.Recv()
		if err == io.EOF {
			log.Println("PushNotifications: поток завершён")
			break
		}
		if err != nil {
			log.Printf("Ошибка получения PushNotifications: %v", err)
			break
		}
		log.Printf("PushNotifications: %s", resp.Message)
	}

	// ================= Client Streaming: AggregatePing =================
	log.Println("=== Client Streaming: AggregatePing ===")
	aggStream, err := client.AggregatePing(context.Background())
	if err != nil {
		log.Fatalf("Ошибка AggregatePing: %v", err)
	}

	for i := 1; i <= 5; i++ {
		msg := "aggregate ping #" + strconv.Itoa(i)
		if err := aggStream.Send(&pb.PingRequest{Message: msg}); err != nil {
			log.Printf("Ошибка отправки AggregatePing: %v", err)
		} else {
			log.Printf("Отправлено AggregatePing: %s", msg)
		}
		time.Sleep(50 * time.Millisecond)
	}

	aggResp, err := aggStream.CloseAndRecv()
	if err != nil {
		log.Printf("Ошибка получения AggregatePing ответа: %v", err)
	} else {
		log.Printf("AggregatePing ответ: %s", aggResp.Message)
	}

	log.Println("=== Все операции завершены успешно ===")
}
