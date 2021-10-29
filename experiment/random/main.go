package main

import (
	"crypto/rand"
	"fmt"
)

func main() {

	p, n := 0, 0

	c := 100
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, v := range b {
		//fmt.Printf("v=%d\n", v)
		if v > 127 {
			p++
		} else {
			n++
		}
	}

	fmt.Printf("Positive %d, negative %d", p, n)
}