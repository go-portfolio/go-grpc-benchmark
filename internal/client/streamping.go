package client

import (
	"context"
	"io"
	"log"
	"strconv"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
)

func StreamPing(client pb.BenchmarkServiceClient, count int) {
	log.Println("=== Bidirectional StreamPing ===")
	stream, err := client.StreamPing(context.Background())
	if err != nil {
		log.Fatalf("Не удалось открыть StreamPing: %v", err)
	}

	// Получение ответов
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				LogDebug("StreamPing: поток завершён сервером")
				return
			}
			if err != nil {
				log.Printf("Ошибка получения StreamPing: %v", err)
				return
			}
			LogDebug("StreamPing response: %s", resp.Message)
		}
	}()

	for i := 1; i <= count; i++ {
		msg := "stream ping #" + strconv.Itoa(i)
		if err := stream.Send(&pb.PingRequest{Message: msg}); err != nil {
			log.Printf("Ошибка отправки StreamPing: %v", err)
		} else {
			LogDebug("Отправлено StreamPing: %s", msg)
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err := stream.CloseSend(); err != nil {
		log.Printf("Ошибка закрытия отправки StreamPing: %v", err)
	}
	time.Sleep(500 * time.Millisecond)
}
