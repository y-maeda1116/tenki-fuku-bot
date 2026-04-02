package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/y-maeda1116/template-go-cross/internal/config"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	_ = cfg

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("Application started. Press Ctrl+C to stop.")

	<-ctx.Done()
	log.Println("Shutting down gracefully...")
}
