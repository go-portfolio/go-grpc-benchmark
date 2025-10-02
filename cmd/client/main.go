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
	// Устанавливаем соединение с gRPC-сервером.
	// grpc.WithInsecure() используется для подключения без TLS (удобно для локальных тестов).
	// В реальных проектах лучше использовать защищённое соединение (TLS).
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close() // закрываем соединение по завершении работы клиента

	// Создаём клиент для BenchmarkService на основе сгенерированного кода из proto.
	// Теперь можно вызывать методы (Ping, Stats, StreamPing) как обычные функции.
	client := pb.NewBenchmarkServiceClient(conn)

	// Параметры нагрузки:
	const requests = 1000   // сколько всего запросов отправим
	const concurrency = 50  // сколько одновременно горутин будут работать

	// sync.WaitGroup используется, чтобы дождаться завершения всех горутин.
	var wg sync.WaitGroup
	wg.Add(concurrency) // указываем, что ожидаем 50 горутин

	// Засекаем время начала бенчмарка.
	start := time.Now()

	// Запускаем 50 параллельных горутин (workers), которые будут слать запросы.
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done() // сообщаем, что горутина завершила работу
			// Каждая горутина отправляет часть запросов: 1000 / 50 = 20
			for j := 0; j < requests/concurrency; j++ {
				// Отправляем RPC-запрос Ping с сообщением "ping"
				_, err := client.Ping(context.Background(), &pb.PingRequest{Message: "ping"})
				if err != nil {
					// Если что-то пошло не так (например, обрыв соединения), логируем ошибку
					log.Printf("Error: %v", err)
				}
			}
		}()
	}

	// Ждём, пока все горутины выполнят свою работу (все 1000 запросов будут отправлены).
	wg.Wait()

	// Засекаем время окончания и вычисляем общее время выполнения.
	elapsed := time.Since(start)

	// Выводим результаты бенчмарка:
	// - сколько всего запросов отправили
	// - за какое время
	log.Printf("Completed %d requests in %s", requests, elapsed)

	// Считаем RPS (Requests Per Second = запросов в секунду)
	// Формула: общее количество запросов / общее время (в секундах)
	log.Printf("RPS: %f", float64(requests)/elapsed.Seconds())
}
