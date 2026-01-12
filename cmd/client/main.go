package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	_ "time/tzdata"

	"github.com/jsirianni/websocket-test/client"
)

func main() {
	host := flag.String("host", "localhost", "WebSocket server host")
	port := flag.Int("port", 3003, "WebSocket server port")
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

	// Start client
	c := client.New(*host, *port)
	if err := c.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Client error: %v", err)
	}
}
