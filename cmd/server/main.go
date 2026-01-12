package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	_ "time/tzdata"

	"github.com/jsirianni/websocket-test/server"
)

func main() {
	host := flag.String("host", "0.0.0.0", "Host to bind the WebSocket server to")
	port := flag.Int("port", 3003, "Port for the WebSocket server")
	metricsPort := flag.Int("metrics-port", 9100, "Port for the Prometheus metrics endpoint")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	// Initialize metrics
	metrics := server.NewMetrics()

	// Start metrics server
	go func() {
		if err := server.StartMetricsServer(ctx, *metricsPort); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// Start WebSocket server
	srv := server.New(*host, *port, metrics)
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
