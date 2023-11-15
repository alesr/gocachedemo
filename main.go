package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/alesr/gocachedemo/client"
	"github.com/alesr/gocachedemo/server"
	"github.com/alesr/httpclient"
	"github.com/dgraph-io/ristretto"
	ristrettostore "github.com/eko/gocache/store/ristretto/v4"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(
			os.Stdout, &slog.HandlerOptions{
				AddSource: true,
			},
		),
	)

	// Initialize the server

	server := server.New(logger, "8080")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Start(shutdownCtx); err != nil {
		log.Fatalln("failed to start server:", err)
	}

	// Initialize the HTTP client
	httpCli := httpclient.New()

	// Initialize the ristretto cache

	cacheConfig := ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	}

	ristrettoCache, err := ristretto.NewCache(&cacheConfig)
	if err != nil {
		log.Fatalln("failed to create cache:", err)
	}

	// Initialize the cache manager

	cacheStore := ristrettostore.NewRistretto(ristrettoCache)

	// Initialize the client
	client := client.New(logger, httpCli, cacheStore, "http://localhost:8080")

	// Chache miss
	res, err := client.GetTest(context.Background())
	if err != nil {
		log.Fatalln("failed to get test:", err)
	}

	fmt.Println("request 1", res)

	// Wait for the cache to be populated
	time.Sleep(time.Second)

	// Cache hit
	res2, err := client.GetTest(context.Background())
	if err != nil {
		log.Fatalln("failed to get test:", err)
	}

	fmt.Println("request 2", res2)

	// Cache hit again
	res3, err := client.GetTest(context.Background())
	if err != nil {
		log.Fatalln("failed to get test:", err)
	}

	fmt.Println("request 3", res3)
}
