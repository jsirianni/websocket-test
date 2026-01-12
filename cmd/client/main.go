package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	_ "time/tzdata"

	"github.com/jsirianni/websocket-test/internal/client"
	"github.com/jsirianni/websocket-test/internal/logger"
	"go.uber.org/zap"
)

func main() {
	host := flag.String("host", "localhost", "WebSocket server host")
	port := flag.Int("port", 3003, "WebSocket server port")
	connections := flag.Int("connections", 1, "Number of concurrent connections")
	useTLS := flag.Bool("tls", false, "Use TLS (wss scheme) for WebSocket connection")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error, dpanic, panic, fatal)")
	flag.Parse()

	if *connections < 1 {
		*connections = 1
	}

	log := logger.MustNewFromString(*logLevel)
	defer log.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info("Received shutdown signal")
		cancel()
	}()

	// Start multiple client instances if needed
	var wg sync.WaitGroup
	errChan := make(chan error, *connections)
	var connIDCounter int
	var connIDMutex sync.Mutex

	// Helper function to start a client connection
	startClient := func(connID int) {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c := client.New(*host, *port, *useTLS, log)
			if err := c.Start(ctx); err != nil && err != context.Canceled {
				errChan <- err
			}
		}(connID)
	}

	// Start initial connections
	for i := 0; i < *connections; i++ {
		connIDMutex.Lock()
		connIDCounter++
		id := connIDCounter
		connIDMutex.Unlock()
		startClient(id)
	}

	// Wait for all connections to finish
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Handle errors and spawn replacements until context cancellation
	for {
		select {
		case err := <-errChan:
			log.Error("Client error, spawning replacement", zap.Error(err))
			connIDMutex.Lock()
			connIDCounter++
			id := connIDCounter
			connIDMutex.Unlock()
			startClient(id)
		case <-ctx.Done():
			<-done
			return
		case <-done:
			return
		}
	}
}
