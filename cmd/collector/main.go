package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aztec-collector/pkg/collector"
	"github.com/aztec-collector/pkg/jsonrpc"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("aztec-collector version %s (built %s)\n", version, buildTime)
		os.Exit(0)
	}

	// Setup logger
	logger := log.New(os.Stdout, "[aztec-collector] ", log.LstdFlags)
	logger.Printf("Starting aztec-collector version %s", version)

	// Load configuration
	config, err := collector.LoadConfig(*configPath)
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	if err := config.Validate(); err != nil {
		logger.Fatalf("Invalid configuration: %v", err)
	}

	logger.Printf("Configuration loaded: protocol=%s, rpc_url=%s", config.Protocol.Name, config.Protocol.RPCURL)

	// Create HTTP client
	httpClient := &http.Client{}

	// Create JSON-RPC client
	rpcClient, err := jsonrpc.New(
		jsonrpc.WithURL(config.Protocol.RPCURL),
		jsonrpc.WithHTTPClient(httpClient),
	)
	if err != nil {
		logger.Fatalf("Failed to create RPC client: %v", err)
	}

	// Create collector
	c, err := collector.New(
		collector.WithConfig(config),
		collector.WithRPCClient(rpcClient),
		collector.WithLogger(logger),
	)
	if err != nil {
		logger.Fatalf("Failed to create collector: %v", err)
	}

	// Setup context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Run collector
	logger.Printf("Starting collector with interval %v", config.Interval)
	if err := c.Run(ctx); err != nil && err != context.Canceled {
		logger.Fatalf("Collector error: %v", err)
	}

	logger.Println("Collector stopped")
}

