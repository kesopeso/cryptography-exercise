package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/kesopeso/cryptography-exercise/internal/server"
	"github.com/kesopeso/cryptography-exercise/internal/store"
)

func main() {
	ctx := context.Background()

	dbURL := "postgresql://postgres:postgres@localhost:5432/apidb?sslmode=disable"
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	statusStore := store.NewPostgresStatusStore(conn)
	handler := server.NewRouter(statusStore)

	addr := ":8090"
	httpServer := server.NewHttpServer(&http.Server{Addr: addr, Handler: handler})

	fmt.Printf("server listening on %s\n", addr)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server shutdown error: %v", err)
	}

	fmt.Println("server closed")
}
