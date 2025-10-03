package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/go-portfolio/go-grpc-benchmark/internal/server"
	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	debug := flag.Bool("debug", false, "Enable debug mode")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	if *debug {
		log.Println("[INFO] Debug mode enabled")
	}
	if *verbose {
		log.Println("[INFO] Verbose logging enabled")
	}

	rand.Seed(time.Now().UnixNano())

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Не удалось слушать порт: %v", err)
	}

	// === TLS ===
	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		log.Fatalf("Ошибка загрузки сертификата сервера: %v", err)
	}

	// Подгружаем CA для проверки клиентов (mTLS)
	caCert, err := ioutil.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatalf("Ошибка чтения CA: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert, // включаем mTLS
		ClientCAs:    caCertPool,
	}

	creds := credentials.NewTLS(tlsConfig)
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	srv := server.NewServer(*debug, *verbose)
	pb.RegisterBenchmarkServiceServer(grpcServer, srv)

	log.Println("[INFO] Сервер запущен на :50051 (TLS)")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
