package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"

	"github.com/go-portfolio/go-grpc-benchmark/internal/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	debug := flag.Bool("debug", false, "Enable debug logs")
	verbose := flag.Bool("verbose", false, "Enable verbose logs")
	flag.Parse()

	client.Debug = *debug
	client.Verbose = *verbose

	if *debug {
		log.Println("Debug mode enabled")
	}
	if *verbose {
		log.Println("Verbose logging enabled")
	}

	// === TLS ===
	cert, err := tls.LoadX509KeyPair("certs/client.crt", "certs/client.key")
	if err != nil {
		log.Fatalf("Ошибка загрузки клиентского сертификата: %v", err)
	}

	caCert, err := ioutil.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatalf("Ошибка чтения CA: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	creds := credentials.NewTLS(tlsConfig)

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}
	defer conn.Close()

	c := client.NewBenchmarkClientWithConn(conn)

	client.UnaryPing(c, 1000, 50)
	client.StreamPing(c, 5)
	client.PushNotifications(c, "start")
	client.AggregatePing(c, 5)
}
