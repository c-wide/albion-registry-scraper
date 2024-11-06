package scraper

import (
	"context"
	"time"

	"github.com/c-wide/albion-registry-scraper/internal/database"
	"github.com/c-wide/albion-registry-scraper/internal/fetcher"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/rs/zerolog"
)

const RE_TICKER_DURATION = 3 * time.Minute
const AN_TICKER_DURATION = 30 * time.Second

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

func (s *Scraper) StartRecentEventTicker(ctx context.Context) {
	go func() {
		s.PerformRecentEventCycle(ctx)

		ticker := time.NewTicker(RE_TICKER_DURATION)
		defer ticker.Stop()

		inProgress := false
		for range ticker.C {
			if inProgress {
				continue
			}

			inProgress = true
			s.PerformRecentEventCycle(ctx)
			inProgress = false
		}
	}()
}

func (s *Scraper) StartAllianceNameTicker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(AN_TICKER_DURATION)
		defer ticker.Stop()

		for range ticker.C {
			s.PerformAllianceNameCycle(ctx)
		}
	}()
}
