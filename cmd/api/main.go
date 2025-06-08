package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/delaram/GoTastic/internal/container"
	"github.com/delaram/GoTastic/pkg/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create container
	ctr, err := container.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}
	defer ctr.Close()

	// Start server
	go func() {
		if err := ctr.Router.Run(fmt.Sprintf(":%s", cfg.Server.Port)); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server exiting")
}
