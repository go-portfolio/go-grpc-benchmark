package main

import (
	"context"
	"errors"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"

	"google.golang.org/grpc"
)

// Структура сервера
type server struct {
	pb.UnimplementedBenchmarkServiceServer

	mu        sync.Mutex
	reqCount  int
	totalTime time.Duration
	failCount int
}

// Симуляция обработки с случайной задержкой и шансом ошибки
func simulateProcessing() (time.Duration, error) {
	delay := time.Duration(1+rand.Intn(5)) * time.Millisecond
	time.Sleep(delay)
	if rand.Float32() < 0.02 { // 2% запросов падают
		return delay, errors.New("simulated server error")
	}
	return delay, nil
}

// Unary RPC: Ping
func (s *server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	// Имитация обработки запроса (задержка и случайная ошибка)
	delay, err := simulateProcessing()

	s.mu.Lock()
	if err != nil {
		s.failCount++
	} else {
		s.reqCount++
		s.totalTime += delay
	}
	s.mu.Unlock()

	if err != nil {
		return nil, err
	}

	// Возвращаем эхо-ответ
	return &pb.PingResponse{Message: req.Message}, nil
}


// Unary RPC: Stats
func (s *server) Stats(ctx context.Context, req *pb.StatsRequest) (*pb.StatsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	avg := 0.0
	if s.reqCount > 0 {
		avg = s.totalTime.Seconds() / float64(s.reqCount)
	}
	return &pb.StatsResponse{
		TotalRequests: int32(s.reqCount),
		AvgLatencySec: avg,
	}, nil
}

// Bidirectional Streaming RPC: StreamPing
func (s *server) StreamPing(stream pb.BenchmarkService_StreamPingServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		delay, err := simulateProcessing()
		time.Sleep(delay)

		msg := "echo: " + req.Message
		if err != nil {
			msg = "error: " + req.Message
		} else {
			s.mu.Lock()
			s.reqCount++
			s.totalTime += delay
			s.mu.Unlock()
		}

		if sendErr := stream.Send(&pb.PingResponse{Message: msg}); sendErr != nil {
			return sendErr
		}
	}
}

// Server Streaming RPC: PushNotifications
func (s *server) PushNotifications(req *pb.PingRequest, stream pb.BenchmarkService_PushNotificationsServer) error {
	for i := 1; i <= 5; i++ {
		time.Sleep(time.Duration(50+rand.Intn(50)) * time.Millisecond)
		msg := req.Message + " #" + strconv.Itoa(i)
		stream.Send(&pb.PingResponse{Message: msg})
	}
	return nil
}

// Client Streaming RPC: AggregatePing
func (s *server) AggregatePing(stream pb.BenchmarkService_AggregatePingServer) error {
	count := 0
	messages := ""
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return stream.SendAndClose(&pb.PingResponse{
					Message: "Aggregated " + strconv.Itoa(count) + " messages: " + messages,
				})
			}
			return err
		}
		count++
		messages += req.Message + " | "

		delay, err := simulateProcessing()
		time.Sleep(delay)
		if err == nil {
			s.mu.Lock()
			s.reqCount++
			s.totalTime += delay
			s.mu.Unlock()
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Не удалось слушать порт: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterBenchmarkServiceServer(s, &server{})
	log.Println("Сервер запущен на :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
