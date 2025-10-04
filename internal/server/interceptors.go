package server

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

func CustomUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	elapsed := time.Since(start)

	if err != nil {
		Error("RPC %s failed: %v (duration: %s)", info.FullMethod, err, elapsed)
	} else {
		Debug("RPC %s succeeded (duration: %s)", info.FullMethod, elapsed)
	}

	return resp, err
}

func CustomStreamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	start := time.Now()
	err := handler(srv, ss)
	elapsed := time.Since(start)

	if err != nil {
		Error("Stream %s failed: %v (duration: %s)", info.FullMethod, err, elapsed)
	} else {
		Debug("Stream %s succeeded (duration: %s)", info.FullMethod, elapsed)
	}

	return err
}
