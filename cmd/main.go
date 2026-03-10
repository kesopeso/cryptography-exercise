package main

import (
	"fmt"

	"github.com/kesopeso/cryptography-exercise/internal/cryptography"
)

func main() {
	err := cryptography.GenerateAndSaveECDSAKey("mykey.pem")
	if err != nil {
		fmt.Printf("failed to generate/save key: %v\n", err)
		return
	}
	fmt.Println("key was generated successfully")
}
