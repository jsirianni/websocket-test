package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	_ "time/tzdata"

	"github.com/jsirianni/websocket-test/client"
)

func main() {
	host := flag.String("host", "localhost", "WebSocket server host")
	port := flag.Int("port", 3003, "WebSocket server port")
	connections := flag.Int("connections", 1, "Number of concurrent connections")
	flag.Parse()

	if *connections < 1 {
		*connections = 1
	}

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

	// Start multiple client instances if needed
	var wg sync.WaitGroup
	errChan := make(chan error, *connections)

	for i := 0; i < *connections; i++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()
			c := client.New(*host, *port)
			if err := c.Start(ctx); err != nil && err != context.Canceled {
				errChan <- err
			}
		}(i + 1)
	}

	// Wait for all connections to finish
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Return first error or wait for context cancellation
	select {
	case err := <-errChan:
		log.Fatalf("Client error: %v", err)
	case <-ctx.Done():
		<-done
	case <-done:
	}
}
