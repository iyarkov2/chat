package main

import (
	"fmt"
	"github.com/iyarkov2/chat/core/entity"
)

func main() {
	foo := Foo {
		Versioned : entity.Versioned {
			Version: 1,
		},
		Name: "Foo 1",
	}
	fmt.Printf("%v\n", foo)

	foo.Name = "Foo 2"

	bar := Bar {}
	bar.Name = "Bar 1"
	bar.Version = 1

	fmt.Printf("%v\n", bar)
}

type Foo struct {
	entity.Versioned
	Name string
}

type Bar struct {
	Foo
	lmt uint64
}