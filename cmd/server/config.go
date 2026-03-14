package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type config struct {
	dbURL       string
	keyPath     string
	addr        string
	authToken   string
	aesPassword string
}

func loadConfig() config {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("can't load .env file: %v", err)
	}

	return config{
		dbURL:       getEnv("DB_URL"),
		keyPath:     getEnv("KEY_PATH"),
		addr:        getEnv("SERVER_ADDRESS"),
		authToken:   getEnv("AUTH_TOKEN"),
		aesPassword: getEnv("AES_PASSWORD"),
	}
}

func getEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		log.Fatalf("missing configuration variable %s", key)
	}
	return value
}
