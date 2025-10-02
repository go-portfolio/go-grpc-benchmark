package server

import (
	"io"
	"math/rand"
	"strconv"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
)

// Bidirectional Streaming: StreamPing
func (s *Server) StreamPing(stream pb.BenchmarkService_StreamPingServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		delay, err := SimulateProcessing()
		time.Sleep(delay)

		msg := "echo: " + req.Message
		if err != nil {
			msg = "error: " + req.Message
		} else {
			s.mu.Lock()
			s.reqCount++
			s.totalTime += delay
			s.mu.Unlock()
		}

		if sendErr := stream.Send(&pb.PingResponse{Message: msg}); sendErr != nil {
			return sendErr
		}
	}
}

// Server Streaming: PushNotifications
func (s *Server) PushNotifications(req *pb.PingRequest, stream pb.BenchmarkService_PushNotificationsServer) error {
	for i := 1; i <= 5; i++ {
		time.Sleep(time.Duration(50+rand.Intn(50)) * time.Millisecond)
		msg := req.Message + " #" + strconv.Itoa(i)
		stream.Send(&pb.PingResponse{Message: msg})
	}
	return nil
}

// Client Streaming: AggregatePing
func (s *Server) AggregatePing(stream pb.BenchmarkService_AggregatePingServer) error {
	count := 0
	messages := ""
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return stream.SendAndClose(&pb.PingResponse{
					Message: "Aggregated " + strconv.Itoa(count) + " messages: " + messages,
				})
			}
			return err
		}

		count++
		messages += req.Message + " | "

		delay, err := SimulateProcessing()
		time.Sleep(delay)
		if err == nil {
			s.mu.Lock()
			s.reqCount++
			s.totalTime += delay
			s.mu.Unlock()
		}
	}
}
