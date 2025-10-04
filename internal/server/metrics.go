package server

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

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

// --- Сбор метрик размера сообщений ---
var (
	RPCRequestSizeBytes  = prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "grpc_rpc_request_size_bytes", Help: "Size of gRPC requests"}, []string{"method"})
	RPCResponseSizeBytes = prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "grpc_rpc_response_size_bytes", Help: "Size of gRPC responses"}, []string{"method"})
)

func init() {
	prometheus.MustRegister(RPCRequestsTotal, RPCLatencySeconds, RPCRequestSizeBytes, RPCResponseSizeBytes)
}

// PrometheusUnaryInterceptor
func PrometheusUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	// размер запроса
	if m, ok := req.(proto.Message); ok {
		RPCRequestSizeBytes.WithLabelValues(info.FullMethod).Observe(float64(proto.Size(m)))
	}

	resp, err := handler(ctx, req)

	// размер ответа
	if m, ok := resp.(proto.Message); ok {
		RPCResponseSizeBytes.WithLabelValues(info.FullMethod).Observe(float64(proto.Size(m)))
	}

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

// PrometheusStreamInterceptor аналогично для stream RPC
func PrometheusStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
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

func StartPrometheusEndpoint(addr string) {
	Info("Метрики Prometheus доступны на %s/metrics", addr)
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(addr, nil); err != nil {
		Error("Ошибка запуска metrics endpoint: %v", err)
	}
}
