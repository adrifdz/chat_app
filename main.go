package main

import (
	"context"
	"log"
	"net"
	"os"
	"sync"

	pb "github.com/adrifdz/chat_app/proto"

	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
)

var grpcLog glog.LoggerV2

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

// Connection is a representation of the fields that constitute a connection to the server
type Connection struct {
	stream pb.Broadcast_CreateStreamServer
	id     string
	active bool
	error  chan error
}

// Server is a slice of connection pointers
type Server struct {
	Connection []*Connection
}

// CreateStream is a function that streams data to the client
func (s *Server) CreateStream(pconn *pb.Connect, stream pb.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream: stream,
		id:     pconn.User.GetId(),
		active: true,
		error:  make(chan error),
	}

	s.Connection = append(s.Connection, conn)

	return <-conn.error
}

// BroadcastMessage is a function that sends messages to the stream
func (s *Server) BroadcastMessage(ctx context.Context, msg *pb.Message) (*pb.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, conn := range s.Connection {
		wait.Add(1)

		go func(msg *pb.Message, conn *Connection) {
			defer wait.Done()

			if conn.active {
				err := conn.stream.Send(msg)
				grpcLog.Info("Sending message to: ", conn.stream)

				if err != nil {
					grpcLog.Errorf("Error with Stream: %v - Error: %v\n", conn.stream, err)
					conn.active = false
					conn.error <- err
				}
			}
		}(msg, conn)
	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
	return &pb.Close{}, nil
}

func main() {
	var connections []*Connection

	server := &Server{connections}

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Server startup error: %v", err)
	}

	grpcLog.Info("Server is now listening on port 8080")

	pb.RegisterBroadcastServer(grpcServer, server)
	grpcServer.Serve(listener)
}
