package main

import (
	"errors"
	"fmt"
)

func main() {
	fmt.Println("Hello")

	root := fmt.Errorf("root error")

	wrapper1 := fmt.Errorf("w1 caused by %w", root)

	wrapper2 := fmt.Errorf("w2 caused by %w", wrapper1)

	fmt.Printf("Error: %s\n", wrapper2)
	fmt.Printf("Error: %v\n", wrapper2)

	u := errors.Unwrap(wrapper2)
	fmt.Printf("Unwrapped 1: %v\n", u)

	w := errors.Unwrap(u)
	fmt.Printf("Unwrapped 2: %v\n", w)

	fmt.Printf("Unwrapped 3: %v\n", errors.Unwrap(w))

	//rootClone := fmt.Errorf("root error")

	fmt.Printf("Is root: %v\n", errors.Is(wrapper2, root))
}
