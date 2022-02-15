package main

import (
	"context"
	"github.com/iyarkov2/chat/server/api"
	"google.golang.org/grpc"
	"log"
)

func main() {
	conn, err := grpc.Dial("localhost:8888", grpc.WithInsecure(), api.WithClientVersion())
	if err != nil {
		log.Fatalln("Connection error:", err)
	}
	defer conn.Close()

	client := api.NewChatServiceClient(conn)

	request := api.ConnectRequest{
		Name: "John Smith",
	}

	response, err := client.Connect(context.Background(), &request)
	if err != nil {
		panic(err)
	}

	log.Println("Response: ", response)
}



