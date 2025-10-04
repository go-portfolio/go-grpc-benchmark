package server

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

// --- Метрики Prometheus ---
var (
	RPCRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_rpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)
	RPCLatencySeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_rpc_latency_seconds",
			Help:    "Latency distribution of gRPC requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)

func init() {
	prometheus.MustRegister(RPCRequestsTotal, RPCLatencySeconds)
}

// --- Unary Interceptor ---
func PrometheusUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	elapsed := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
		Error("RPC %s failed: %v (duration: %.3f s)", info.FullMethod, err, elapsed)
	} else {
		Debug("RPC %s succeeded (duration: %.3f s)", info.FullMethod, elapsed)
	}

	RPCRequestsTotal.WithLabelValues(info.FullMethod, status).Inc()
	RPCLatencySeconds.WithLabelValues(info.FullMethod).Observe(elapsed)

	return resp, err
}

// --- Stream Interceptor ---
func PrometheusStreamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	start := time.Now()
	err := handler(srv, ss)
	elapsed := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
		Error("Stream %s failed: %v (duration: %.3f s)", info.FullMethod, err, elapsed)
	} else {
		Debug("Stream %s succeeded (duration: %.3f s)", info.FullMethod, elapsed)
	}

	RPCRequestsTotal.WithLabelValues(info.FullMethod, status).Inc()
	RPCLatencySeconds.WithLabelValues(info.FullMethod).Observe(elapsed)

	return err
}

// --- HTTP сервер для метрик ---
func StartPrometheusEndpoint(addr string) {
	Info("Метрики Prometheus доступны на %s/metrics", addr)
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(addr, nil); err != nil {
		Error("Ошибка запуска metrics endpoint: %v", err)
	}
}
