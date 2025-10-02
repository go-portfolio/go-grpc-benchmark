package server

import (
	"io"
	"log"
	"math/rand"
	"strconv"
	"time"

	pb "github.com/go-portfolio/go-grpc-benchmark/proto"
)

// Bidirectional Streaming RPC: StreamPing
func (s *Server) StreamPing(stream pb.BenchmarkService_StreamPingServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			if err != io.EOF {
				log.Printf("StreamPing Recv error: %v", err)
			}
			return err
		}

		delay, procErr := SimulateProcessing()
		time.Sleep(delay)

		msg := req.Message
		if procErr != nil {
			msg = "error: " + msg
		} else {
			s.mu.Lock()
			s.reqCount++
			s.totalTime += delay
			s.mu.Unlock()
		}

		if sendErr := stream.Send(&pb.PingResponse{Message: "echo: " + msg}); sendErr != nil {
			return sendErr
		}
		s.logDebug("StreamPing processed: %s", msg)
	}
}

// Server Streaming RPC: PushNotifications
func (s *Server) PushNotifications(req *pb.PingRequest, stream pb.BenchmarkService_PushNotificationsServer) error {
	for i := 1; i <= 5; i++ {
		time.Sleep(time.Duration(50+rand.Intn(50)) * time.Millisecond)
		msg := req.Message + " #" + strconv.Itoa(i)
		if err := stream.Send(&pb.PingResponse{Message: msg}); err != nil {
			return err
		}
		s.logDebug("PushNotifications sent: %s", msg)
	}
	return nil
}

// Client Streaming RPC: AggregatePing
func (s *Server) AggregatePing(stream pb.BenchmarkService_AggregatePingServer) error {
	count := 0
	messages := ""
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				response := "Aggregated " + strconv.Itoa(count) + " messages: " + messages
				s.logDebug("AggregatePing done: %s", response)
				return stream.SendAndClose(&pb.PingResponse{Message: response})
			}
			return err
		}
		count++
		messages += req.Message + " | "

		delay, procErr := SimulateProcessing()
		time.Sleep(delay)
		if procErr == nil {
			s.mu.Lock()
			s.reqCount++
			s.totalTime += delay
			s.mu.Unlock()
		}
		s.logDebug("AggregatePing received: %s", req.Message)
	}
}
