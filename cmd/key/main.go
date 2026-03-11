package main

import (
	"fmt"
	"log"

	"github.com/kesopeso/cryptography-exercise/internal/cryptography"
)

func main() {
	err := cryptography.GenerateAndSaveECDSAKey("server.pem")
	if err != nil {
		log.Fatalf("failed generating key: %v", err)
	}

	fmt.Println("key generated")
}
