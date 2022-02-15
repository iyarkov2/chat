package main

import (
	"fmt"
	"log"
	"runtime"
)

func main() {
	fmt.Println(runtime.GOMAXPROCS(0))

	log.Fatal(nil)
}