package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kesopeso/cryptography-exercise/internal/server"
)

func main() {
	handler := server.NewRouter()
	addr := ":8090"
	httpServer := server.NewHttpServer(&http.Server{Addr: addr, Handler: handler})

	fmt.Printf("server listening on %s\n", addr)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server shutdown error: %v", err)
	}

	fmt.Println("server closed")
}
