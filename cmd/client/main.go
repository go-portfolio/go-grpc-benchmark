package main

import (
	"flag"
	"log"

	"github.com/go-portfolio/go-grpc-benchmark/internal/client"
)

func main() {
	debug := flag.Bool("debug", false, "Enable debug logs")
	verbose := flag.Bool("verbose", false, "Enable verbose logs for RPC calls")
	flag.Parse()

	client.Debug = *debug
	client.Verbose = *verbose

	if *debug {
		log.Println("Debug mode enabled")
	}
	if *verbose {
		log.Println("Verbose logging enabled")
	}

	c := client.NewBenchmarkClient("localhost:50051")
	client.UnaryPing(c, 1000, 50)
}
