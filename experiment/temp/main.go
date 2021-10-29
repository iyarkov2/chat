package main

import (
	"fmt"
	"github.com/google/uuid"
)

func main() {

	uuid := uuid.New()
	fmt.Printf("UUID: %s, length: %d, length S: %d\n", uuid, len(uuid), len(uuid.String()))

	optPar("ABC")
	optPar("ABC", "a", "b")
	optPar("ABC", "a", "b", "c")
	optPar("ABC", "d")
}

func optPar(s string, o...string) {
	fmt.Printf("param s: %s\n", s)
	for idx, val := range o {
		fmt.Printf("idx: %d val: %s\n", idx, val)
	}
}