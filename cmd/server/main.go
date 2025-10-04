package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/go-portfolio/go-grpc-benchmark/internal/server"
	pb "github.com/go-portfolio/go-grpc-benchmark/proto"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	// Создаём Jaeger экспортер и сразу используем его
	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")),
	)
	if err != nil {
		return nil, err
	}

	// TracerProvider использует экспортер
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp), // <-- здесь exp используется
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("grpc-benchmark-server"),
		)),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}

// setupFlags парсит флаги командной строки
func setupFlags() (debug, verbose *bool, metricsPort *string) {
	debug = flag.Bool("debug", false, "Enable debug mode")
	verbose = flag.Bool("verbose", false, "Enable verbose logging")
	metricsPort = flag.String("metrics-port", ":9090", "Prometheus metrics endpoint")
	flag.Parse()
	return
}

func main() {
	// ------------------------------
	// Инициализация логгера
	// ------------------------------
	if err := server.InitLogger("../../logs/server.log"); err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err)
	}
	defer server.CloseLogger()

	server.IsDebug = true
	server.Verbose = true

	// ------------------------------
	// Флаги командной строки
	// ------------------------------
	debug, verbose, metricsPort := setupFlags()
	if *debug {
		server.Info("Debug mode enabled")
	}
	if *verbose {
		server.Info("Verbose logging enabled")
	}

	// ------------------------------
	// Инициализация OpenTelemetry
	// ------------------------------
	tp, err := initTracer()
	if err != nil {
		server.Error("Ошибка инициализации OpenTelemetry: %v", err)
		os.Exit(1)
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	// ------------------------------
	// Случайные задержки для тестов
	// ------------------------------
	rand.Seed(time.Now().UnixNano())

	// ------------------------------
	// TCP Listener
	// ------------------------------
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		server.Error("Не удалось слушать порт: %v", err)
		os.Exit(1)
	}

	// ------------------------------
	// Настройка TLS
	// ------------------------------
	tlsConfig, err := server.LoadTLS("certs/server.crt", "certs/server.key", "certs/ca.crt")
	if err != nil {
		server.Error("Ошибка TLS: %v", err)
		os.Exit(1)
	}
	creds := credentials.NewTLS(tlsConfig)

	grpc_prometheus.EnableHandlingTimeHistogram()

	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor, // сначала Prometheus
			server.PrometheusUnaryInterceptor,      // ваш кастомный, если нужен
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor, // сначала Prometheus
			server.PrometheusStreamInterceptor,      // ваш кастомный
		),
		grpc.StatsHandler(otelgrpc.NewServerHandler()), // OpenTelemetry в конце
	)

	srv := server.NewServer(*debug, *verbose)
	pb.RegisterBenchmarkServiceServer(grpcServer, srv)
	grpc_prometheus.Register(grpcServer)

	// ------------------------------
	// Запуск Prometheus метрик
	// ------------------------------
	go server.StartPrometheusEndpoint(*metricsPort)

	// ------------------------------
	// Запуск gRPC сервера
	// ------------------------------
	server.Info("Сервер запущен на :50051 (TLS + Prometheus)")
	if err := grpcServer.Serve(lis); err != nil {
		server.Error("Ошибка сервера: %v", err)
	}
}
