package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type config struct {
	dbURL   string
	keyPath string
	addr    string
}

func loadConfig() config {
	if err := godotenv.Load(); err != nil {
		panic(fmt.Sprintf("can't load .env file: %v", err))
	}

	return config{
		dbURL:   getEnv("DB_URL"),
		keyPath: getEnv("KEY_PATH"),
		addr:    getEnv("SERVER_ADDRESS"),
	}
}

func getEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		panic(fmt.Sprintf("missing configuration variable %s", key))
	}
	return value
}
