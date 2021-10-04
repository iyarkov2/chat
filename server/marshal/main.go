package main

import (
	"fmt"
	"io/ioutil"
	"log"

	_ "github.com/iyarkov2/chat/api"
	"google.golang.org/protobuf/proto"

	"github.com/iyarkov2/chat/server/api"
)

func main() {
	msg := &api.ConnectRequest{
		Name : "Vasya",
	}

	out, err := proto.Marshal(msg)

	if err != nil {
		log.Fatalln("Can not marshall message:", err)
	}

	if err := ioutil.WriteFile("../out/message.out", out, 0644); err != nil {
		log.Fatalln("Failed to write address book:", err)
	}
	fmt.Printf("Done\n")
}