package server

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartPrometheusEndpoint(port string) {
	go func() {
		log.Printf("[INFO] Метрики доступны на %s/metrics", port)
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(port, nil); err != nil {
			log.Fatalf("Ошибка запуска metrics endpoint: %v", err)
		}
	}()
}
