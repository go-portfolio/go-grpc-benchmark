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
	// Подключаемся к gRPC-серверу на localhost:50051
	// grpc.WithInsecure() используется для локальных тестов без TLS
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Не удалось подключиться к серверу: %v", err)
	}
	defer conn.Close() // Закрываем соединение после завершения работы клиента

	// Создаём клиент для сервиса BenchmarkService
	client := pb.NewBenchmarkServiceClient(conn)

	// ---------------- Часть 1: Бенчмарк обычного Ping ----------------
	const requests = 1000   // общее количество Ping-запросов
	const concurrency = 50  // количество параллельных горутин (потоков)

	log.Printf("Начинаем отправку %d Ping-запросов с concurrency=%d", requests, concurrency)

	var wg sync.WaitGroup
	wg.Add(concurrency) // указываем, что ждём завершения 50 горутин

	start := time.Now() // фиксируем время начала бенчмарка

	// Запускаем несколько горутин для параллельной отправки Ping-запросов
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done() // уведомляем WaitGroup о завершении горутины
			for j := 0; j < requests/concurrency; j++ {
				// Отправка Ping-запроса на сервер
				_, err := client.Ping(context.Background(), &pb.PingRequest{Message: "ping"})
				if err != nil {
					log.Printf("Ошибка при Ping: %v", err)
				}
			}
		}()
	}

	wg.Wait() // Ждём завершения всех горутин
	elapsed := time.Since(start) // вычисляем общее время выполнения

	// Выводим результаты бенчмарка
	log.Printf("Бенчмарк Ping завершён.")
	log.Printf("Всего отправлено запросов: %d", requests)
	log.Printf("Общее время выполнения: %s", elapsed)
	log.Printf("Средняя скорость (RPS): %.2f запросов в секунду", float64(requests)/elapsed.Seconds())

	// ---------------- Часть 2: Опрос статистики сервера ----------------
	// Отправляем RPC-запрос Stats для получения метрик сервера
	statsResp, err := client.Stats(context.Background(), &pb.StatsRequest{})
	if err != nil {
		log.Fatalf("Не удалось получить статистику сервера: %v", err)
	}

	// Выводим статистику сервера
	log.Println("Статистика сервера:")
	log.Printf("Общее количество обработанных сервером запросов: %d", statsResp.TotalRequests)
	log.Printf("Средняя задержка обработки одного запроса на сервере: %.6f сек", statsResp.AvgLatencySec)

	// ---------------- Часть 3: Двунаправленный стриминг StreamPing ----------------
	// Создаём стрим для двунаправленного общения с сервером
	stream, err := client.StreamPing(context.Background())
	if err != nil {
		log.Fatalf("Не удалось открыть StreamPing: %v", err)
	}

	log.Println("Начинаем двунаправленный стриминг StreamPing (отправляем 5 сообщений)")

	// Горутина для приёма сообщений от сервера
	go func() {
		for {
			// Чтение ответа от сервера
			resp, err := stream.Recv()
			if err == io.EOF {
				// поток завершился корректно
				log.Println("StreamPing: сервер завершил поток.")
				return
			}
			if err != nil {
				log.Printf("Ошибка при получении сообщения из StreamPing: %v", err)
				return
			}
			log.Printf("Ответ от сервера по StreamPing: %s", resp.Message)
		}
	}()

	// Отправляем 5 сообщений через стрим
	for i := 1; i <= 5; i++ {
		// Формируем сообщение с номером
		msg := "stream ping #" + strconv.Itoa(i) // конвертируем число в строку
		if err := stream.Send(&pb.PingRequest{Message: msg}); err != nil {
			log.Printf("Ошибка при отправке сообщения через StreamPing: %v", err)
		} else {
			log.Printf("Отправлено сообщение через StreamPing: %s", msg)
		}
		time.Sleep(100 * time.Millisecond) // небольшой таймаут для наглядности
	}

	// Закрываем отправку сообщений (клиент больше не будет отправлять)
	if err := stream.CloseSend(); err != nil {
		log.Printf("Ошибка при закрытии отправки StreamPing: %v", err)
	}

	// Даём время горутине на приём всех ответов
	time.Sleep(500 * time.Millisecond)

	log.Println("Двунаправленный стриминг StreamPing завершён.")
	log.Println("Все операции завершены успешно.")
}
