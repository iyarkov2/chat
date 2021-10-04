package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/iyarkov2/chat/server/api"
)

type chatServer struct {
	api.UnimplementedChatServiceServer
}

func (s *chatServer) Connect(ctx context.Context, request *api.ConnectRequest) (*api.ConnectResponse, error) {
	fmt.Printf("Received request [%v]\n", request)
	result := &api.ConnectResponse{
		Status : api.ConnectResponse_SUCCESS,
		UserId: 1,
	}
	return result, nil
}

func (s *chatServer) Post(stream api.ChatService_PostServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Printf("Received: %v", in)

		msg := &api.PostResponse{
			Id: 1,
			UserId: 2,
			Text: "Response: " + in.Text,
		}
		stream.Send(msg)
	}
}

func newServer() *chatServer {
	s := &chatServer{}
	return s
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8888))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	api.RegisterChatServiceServer(grpcServer, newServer())
	err2 := grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err2)
	}
}
