package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	_ "time/tzdata"

	"github.com/jsirianni/websocket-test/internal/logger"
	"github.com/jsirianni/websocket-test/internal/server"
	"go.uber.org/zap"
)

func main() {
	host := flag.String("host", "0.0.0.0", "Host to bind the WebSocket server to")
	port := flag.Int("port", 3003, "Port for the WebSocket server")
	metricsPort := flag.Int("metrics-port", 9100, "Port for the Prometheus metrics endpoint")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error, dpanic, panic, fatal)")
	flag.Parse()

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

	// Initialize metrics
	metrics := server.NewMetrics()

	// Start metrics server
	go func() {
		if err := server.StartMetricsServer(ctx, *metricsPort, log); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Metrics server error", zap.Error(err))
		}
	}()

	// Start WebSocket server
	srv := server.New(*host, *port, metrics, log)
	if err := srv.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("Server error", zap.Error(err))
	}
}
