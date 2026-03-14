package main

import (
	"fmt"
	"os"

	"github.com/kesopeso/cryptography-exercise/pkg/cryptography"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	err := cryptography.GenerateAndSaveECDSAKey("server.pem")
	if err != nil {
		return fmt.Errorf("failed generating key: %v", err)
	}

	fmt.Println("key generated")
	return nil
}
