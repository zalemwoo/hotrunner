package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	fmt.Printf("argv is : %v\n", os.Args)
	addr := os.Args[1]
	fmt.Printf("server is listen at: %s\n start", addr)
	http.ListenAndServe(os.Args[1], nil)
}
