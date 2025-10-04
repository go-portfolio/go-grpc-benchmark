package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus" // interceptors для сбора метрик Prometheus
	"github.com/prometheus/client_golang/prometheus/promhttp"      // HTTP handler для метрик

	"github.com/go-portfolio/go-grpc-benchmark/internal/server" // твой gRPC сервер
	pb "github.com/go-portfolio/go-grpc-benchmark/proto"        // protobuf генерация
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	if err := server.InitLogger("../../logs/server.log"); err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err)
	}
	defer server.CloseLogger()

	server.IsDebug = true
	server.Verbose = true

	// === Настройка логирования ===
	// log.SetFlags определяет формат вывода: дата, время, микросекунды
	// log.SetOutput(os.Stdout) гарантирует, что вывод будет идти в терминал без буферизации
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(os.Stdout)

	// === Парсинг флагов запуска ===
	debug := flag.Bool("debug", false, "Enable debug mode")                            // включение debug-логов
	verbose := flag.Bool("verbose", false, "Enable verbose logging")                   // включение verbose-логов
	metricsPort := flag.String("metrics-port", ":9090", "Prometheus metrics endpoint") // порт для HTTP сервера метрик
	flag.Parse()

	// Логирование состояния режимов
	if *debug {
		log.Println("[INFO] Debug mode enabled")
	}
	if *verbose {
		log.Println("[INFO] Verbose logging enabled")
	}

	// === Инициализация генератора случайных чисел ===
	// Нужно для нагрузочного тестирования, имитации задержек и случайных данных
	rand.Seed(time.Now().UnixNano())

	// === Создаем TCP-сокет для gRPC сервера ===
	// Порт 50051 — стандартный для gRPC (можно изменить)
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Не удалось слушать порт: %v", err)
	}

	// === Настройка TLS/mTLS ===
	// Загружаем сертификат сервера и приватный ключ
	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		log.Fatalf("Ошибка загрузки сертификата сервера: %v", err)
	}

	// Загружаем CA для проверки клиентских сертификатов (mutual TLS)
	caCert, err := ioutil.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatalf("Ошибка чтения CA: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Конфигурация TLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},        // серверный сертификат
		ClientAuth:   tls.RequireAndVerifyClientCert, // требовать клиентский сертификат
		ClientCAs:    caCertPool,                     // доверенные CA для проверки клиента
	}

	// Создаем gRPC credentials
	creds := credentials.NewTLS(tlsConfig)

	// === Создаем gRPC сервер с Prometheus метриками ===
	// Unary и Stream interceptors автоматически собирают метрики для всех RPC
	grpcServer := grpc.NewServer(
		grpc.Creds(creds), // включаем TLS/mTLS
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),   // метрики для unary RPC
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor), // метрики для stream RPC
	)

	// === Регистрируем свой gRPC сервис ===
	srv := server.NewServer(*debug, *verbose) // передаем флаги debug/verbose внутрь сервера
	pb.RegisterBenchmarkServiceServer(grpcServer, srv)

	// Регистрируем стандартные метрики Prometheus для gRPC
	grpc_prometheus.Register(grpcServer)

	// === Запуск HTTP сервера для Prometheus ===
	// Отдает метрики по адресу http://<host>:9090/metrics
	go func() {
		log.Printf("[INFO] Метрики доступны на %s/metrics", *metricsPort)
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(*metricsPort, nil); err != nil {
			log.Fatalf("Ошибка запуска metrics endpoint: %v", err)
		}
	}()

	// === Запуск gRPC сервера ===
	log.Println("[INFO] Сервер запущен на :50051 (TLS + Prometheus)")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
