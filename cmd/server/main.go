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
	cfg := loadConfig()
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, cfg.dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	statusStore := store.NewPostgresStatusStore(conn)
	handler := server.NewRouter(statusStore, cfg.keyPath)

	httpServer := server.NewHttpServer(&http.Server{Addr: cfg.addr, Handler: handler})

	fmt.Printf("server listening on %s\n", cfg.addr)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server shutdown error: %v", err)
	}

	fmt.Println("server closed")
}
