package main

import (
	"context"
	"flag"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
	"google.golang.org/grpc"
)

var debug bool
var verbose bool

func logDebug(format string, v ...interface{}) {
	if debug {
		log.Printf("[DEBUG] "+format, v...)
	}
}

func logVerbose(format string, v ...interface{}) {
	if verbose {
		log.Printf("[VERBOSE] "+format, v...)
	}
}

func main() {
	// Флаг включения режима отладки
	flag.BoolVar(&debug, "debug", false, "Enable debug logs")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logs for RPC calls")
	flag.Parse()

	log.Println("=== Запуск клиента gRPC Benchmark ===")
	if debug {
		log.Println("Debug mode enabled")
	}
	if verbose {
		log.Println("Verbose logging enabled")
	}

	// ---------------- Подключение к серверу ----------------
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Не удалось подключиться к серверу: %v", err)
	}
	defer conn.Close()

	client := pb.NewBenchmarkServiceClient(conn)

	// ================= Unary RPC: Ping =================
	const requests = 1000
	const concurrency = 50

	log.Printf("=== Бенчмарк Unary Ping ===")
	var wg sync.WaitGroup
	wg.Add(concurrency)
	start := time.Now()

	var successCount int64
	var failCount int64
	var mu sync.Mutex

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
					logDebug("Worker %d: Ping error: %v", workerID, err)
					logVerbose("Ping failed: %v", err)
				} else {
					successCount++
					logDebug("Worker %d: Ping response: %s", workerID, resp.Message)
					logVerbose("Ping succeeded")
				}
				mu.Unlock()
			}
		}(i)
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

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				logDebug("StreamPing: поток завершён сервером")
				return
			}
			if err != nil {
				log.Printf("Ошибка получения StreamPing: %v", err)
				return
			}
			logDebug("StreamPing response: %s", resp.Message)
		}
	}()

	for i := 1; i <= 5; i++ {
		msg := "stream ping #" + strconv.Itoa(i)
		if err := stream.Send(&pb.PingRequest{Message: msg}); err != nil {
			log.Printf("Ошибка отправки StreamPing: %v", err)
		} else {
			logDebug("Отправлено StreamPing: %s", msg)
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err := stream.CloseSend(); err != nil {
		log.Printf("Ошибка закрытия отправки StreamPing: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// ================= Server Streaming: PushNotifications =================
	log.Println("=== Server Streaming: PushNotifications ===")
	notifyStream, err := client.PushNotifications(context.Background(), &pb.PingRequest{Message: "start"})
	if err != nil {
		log.Fatalf("Ошибка PushNotifications: %v", err)
	}
	for {
		resp, err := notifyStream.Recv()
		if err == io.EOF {
			logDebug("PushNotifications: поток завершён")
			break
		}
		if err != nil {
			log.Printf("Ошибка получения PushNotifications: %v", err)
			break
		}
		logDebug("PushNotifications response: %s", resp.Message)
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
			logDebug("Отправлено AggregatePing: %s", msg)
		}
		time.Sleep(50 * time.Millisecond)
	}

	aggResp, err := aggStream.CloseAndRecv()
	if err != nil {
		log.Printf("Ошибка получения AggregatePing ответа: %v", err)
	} else {
		logDebug("AggregatePing ответ: %s", aggResp.Message)
	}

	log.Println("=== Все операции завершены успешно ===")
}
