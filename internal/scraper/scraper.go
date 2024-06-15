package scraper

import (
	"context"
	"time"

	"github.com/c-wide/albion-registry-scraper/internal/database"
	"github.com/c-wide/albion-registry-scraper/internal/fetcher"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/rs/zerolog"
)

const DEFAULT_TICKER_DURATION = 3 * time.Minute

type Scraper struct {
	logger      zerolog.Logger
	queries     *database.Queries
	eventCaches map[string]*lru.Cache[uint64, any]
	fetchers    map[string]*fetcher.Fetcher
}

func New(
	logger zerolog.Logger,
	queries *database.Queries,
	eventCaches map[string]*lru.Cache[uint64, any],
	fetchers map[string]*fetcher.Fetcher,
) *Scraper {
	return &Scraper{
		logger:      logger,
		queries:     queries,
		eventCaches: eventCaches,
		fetchers:    fetchers,
	}
}

func (s *Scraper) StartTicker(ctx context.Context) {
	s.PerformCycle(ctx)

	ticker := time.NewTicker(DEFAULT_TICKER_DURATION)
	defer ticker.Stop()

	for range ticker.C {
		s.PerformCycle(ctx)
	}
}
