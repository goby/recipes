package main

import (
	"fmt"
	"net/http"

	"golang.org/x/net/http2"
)

func main() {
	fmt.Println("vim-go")
	server := http.Server{}
	http2.VerboseLogs = true
}
