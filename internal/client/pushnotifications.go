package client

import (
	"context"
	"io"
	"log"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
)

func PushNotifications(client pb.BenchmarkServiceClient, startMsg string) {
	log.Println("=== Server Streaming: PushNotifications ===")
	stream, err := client.PushNotifications(context.Background(), &pb.PingRequest{Message: startMsg})
	if err != nil {
		log.Fatalf("Ошибка PushNotifications: %v", err)
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			LogDebug("PushNotifications: поток завершён")
			break
		}
		if err != nil {
			log.Printf("Ошибка получения PushNotifications: %v", err)
			break
		}
		LogDebug("PushNotifications response: %s", resp.Message)
	}
}
