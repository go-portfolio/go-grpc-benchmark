package main

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"

	"google.golang.org/grpc"
)

// server реализует интерфейс BenchmarkService, сгенерированный из .proto.
// Здесь мы встраиваем "UnimplementedBenchmarkServiceServer", чтобы
// автоматически удовлетворять интерфейсу (и быть совместимыми с будущими
// изменениями в proto).
type server struct {
	pb.UnimplementedBenchmarkServiceServer

	// Для подсчёта статистики по запросам
	mu        sync.Mutex   // мьютекс для защиты разделяемых данных
	reqCount  int          // количество обработанных запросов
	totalTime time.Duration // суммарное время обработки запросов
}

// Ping — простой RPC метод (unary RPC).
// Клиент отправляет одно сообщение, сервер возвращает одно сообщение.
// Здесь метод просто возвращает обратно то, что прислал клиент ("эхо").
func (s *server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	// фиксируем начало обработки запроса
	start := time.Now()

	// создаём ответ (эхо-ответ)
	resp := &pb.PingResponse{Message: req.Message}

	// измеряем время выполнения (очень быстро, т.к. логики почти нет)
	elapsed := time.Since(start)

	// обновляем статистику с защитой мьютексом
	s.mu.Lock()
	s.reqCount++
	s.totalTime += elapsed
	s.mu.Unlock()

	// возвращаем ответ клиенту
	return resp, nil
}

// Stats — RPC метод для получения статистики работы сервера.
// Клиент может запросить, сколько запросов было выполнено и среднюю задержку.
func (s *server) Stats(ctx context.Context, req *pb.StatsRequest) (*pb.StatsResponse, error) {
	// защищаем доступ к данным мьютексом
	s.mu.Lock()
	defer s.mu.Unlock()

	var avg float64
	if s.reqCount > 0 {
		// среднее время в секундах
		avg = s.totalTime.Seconds() / float64(s.reqCount)
	}

	// формируем ответ с метриками
	return &pb.StatsResponse{
		TotalRequests: int32(s.reqCount),
		AvgLatencySec: avg,
	}, nil
}

// StreamPing — пример двунаправленного стриминга (bidirectional streaming).
// Клиент открывает поток и может отправлять в него несколько сообщений подряд.
// Сервер читает входящие сообщения и сразу отправляет ответ обратно (эхо).
// Таким образом можно построить "чат" или "пинг-понг".
func (s *server) StreamPing(stream pb.BenchmarkService_StreamPingServer) error {
	for {
		// ждём следующее сообщение от клиента
		req, err := stream.Recv()
		if err != nil {
			// если клиент закрыл поток или произошла ошибка — выходим
			return err
		}

		// сразу же отправляем ответ обратно
		if err := stream.Send(&pb.PingResponse{Message: "echo: " + req.Message}); err != nil {
			return err
		}
	}
}

func main() {
	// Создаём TCP-листенер на порту 50051.
	// Все gRPC соединения будут приходить сюда.
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Создаём новый gRPC сервер.
	s := grpc.NewServer()

	// Регистрируем наш сервис BenchmarkService на сервере.
	// Теперь сервер знает, как обрабатывать RPC вызовы от клиентов.
	pb.RegisterBenchmarkServiceServer(s, &server{})

	log.Println("Server listening on :50051")

	// Запускаем сервер (метод Serve блокирует текущую горутину
	// и принимает соединения до остановки приложения).
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
