package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	adapter "github.com/axiomhq/axiom-go/adapters/zerolog"
	"github.com/c-wide/albion-registry-scraper/config"
	"github.com/c-wide/albion-registry-scraper/internal/database"
	"github.com/c-wide/albion-registry-scraper/internal/fetcher"
	"github.com/c-wide/albion-registry-scraper/internal/scraper"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Unable to load environment variables from .env file")
	}

	// Create default logger
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	// If Axiom logging is enabled, create new logger
	_, axiomEnabled := os.LookupEnv("AXIOM_TOKEN")
	if axiomEnabled {
		writer, err := adapter.New()
		if err != nil {
			logger.Fatal().Err(err).Msg("Unable to create Axiom writer")
		}

		defer writer.Close()

		logger = zerolog.New(io.MultiWriter(writer, os.Stderr)).With().Timestamp().Logger()
	}

	// Database stuff
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Fatal().Err(err).Msg("Unable to create database pool")
	}
	defer pool.Close()

	queries := database.New(pool)

	// Create event caches and fetchers
	var eventCaches = make(map[string]*lru.Cache[uint64, any])
	var fetchers = make(map[string]*fetcher.Fetcher)

	for region, url := range config.ServerURLs {
		cache, err := lru.New[uint64, any](2500)
		if err != nil {
			logger.Fatal().Err(err).Str("region", region).Msg("Unable to create LRU cache")
		}

		fetcher := fetcher.New(url)

		eventCaches[region] = cache
		fetchers[region] = fetcher
	}

	// Create SIGINT context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Create scraper and start tickers
	s := scraper.New(logger, queries, eventCaches, fetchers)
	s.StartRecentEventTicker(ctx)
	s.StartAllianceNameTicker(ctx)

	// Wait for SIGINT signal
	<-ctx.Done()
}
