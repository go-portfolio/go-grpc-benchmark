package main

import (
	"context"
	"io"
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// server реализует интерфейс BenchmarkService и хранит статистику всех типов RPC
type server struct {
	pb.UnimplementedBenchmarkServiceServer

	mu          sync.Mutex
	reqCount    int
	streamCount int
	totalTime   time.Duration
}

// ---------------- Unary RPC: Ping ----------------
// Клиент отправляет одно сообщение, сервер отвечает одним сообщением
func (s *server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	start := time.Now()

	// Можно использовать метаданные для логирования или аутентификации
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		log.Printf("Unary Ping received metadata: %v", md)
	}

	resp := &pb.PingResponse{Message: req.Message}

	// Обновляем статистику
	elapsed := time.Since(start)
	s.mu.Lock()
	s.reqCount++
	s.totalTime += elapsed
	s.mu.Unlock()

	return resp, nil
}

// ---------------- Unary RPC: Stats ----------------
// Возвращает статистику по всем вызовам
func (s *server) Stats(ctx context.Context, req *pb.StatsRequest) (*pb.StatsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var avg float64
	if s.reqCount > 0 {
		avg = s.totalTime.Seconds() / float64(s.reqCount)
	}

	return &pb.StatsResponse{
		TotalRequests: int32(s.reqCount),
		AvgLatencySec: avg,
	}, nil
}

// ---------------- Bidirectional Streaming: StreamPing ----------------
// Клиент отправляет поток запросов, сервер отвечает потоком
func (s *server) StreamPing(stream pb.BenchmarkService_StreamPingServer) error {
	s.mu.Lock()
	s.streamCount++
	s.mu.Unlock()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// клиент завершил поток
			return nil
		}
		if err != nil {
			return err
		}

		// Эхо-сообщение
		resp := &pb.PingResponse{Message: "echo: " + req.Message}

		// Можно имитировать задержку для тестов
		time.Sleep(10 * time.Millisecond)

		if err := stream.Send(resp); err != nil {
			return err
		}

		// Можно фиксировать статистику стримов
		s.mu.Lock()
		s.totalTime += 10 * time.Millisecond
		s.mu.Unlock()
	}
}

// ---------------- Server Streaming: PushNotifications ----------------
// Сервер сам инициирует поток сообщений (например, уведомления)
func (s *server) PushNotifications(req *pb.PingRequest, stream pb.BenchmarkService_PushNotificationsServer) error {
	for i := 1; i <= 5; i++ {
		msg := &pb.PingResponse{Message: "notification #" + string(i)}
		if err := stream.Send(msg); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// ---------------- Client Streaming: AggregatePing ----------------
// Клиент отправляет поток сообщений, сервер возвращает один агрегированный ответ
func (s *server) AggregatePing(stream pb.BenchmarkService_AggregatePingServer) error {
	var count int
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			// клиент завершил поток — возвращаем агрегированный результат
			return stream.SendAndClose(&pb.PingResponse{
				Message: "Received " + string(count) + " messages",
			})
		}
		if err != nil {
			return err
		}
		count++
	}
}

// ---------------- Main ----------------
func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Не удалось открыть порт: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterBenchmarkServiceServer(s, &server{})

	log.Println("Server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
