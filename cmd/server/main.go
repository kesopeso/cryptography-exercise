package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/kesopeso/cryptography-exercise/internal/server"
	"github.com/kesopeso/cryptography-exercise/internal/service"
	"github.com/kesopeso/cryptography-exercise/internal/store"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := loadConfig()
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, cfg.dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	statusStore := store.NewPostgresStatusStore(conn)
	statusService := service.NewDefaultStatusService(statusStore, cfg.aesPassword)
	handler := server.NewRouter(statusService, cfg.keyPath, cfg.authToken)

	httpServer := server.NewHttpServer(&http.Server{Addr: cfg.addr, Handler: handler})

	fmt.Println("server listening on", cfg.addr)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server shutdown error: %v", err)
	}

	fmt.Println("server closed")
	return nil
}
