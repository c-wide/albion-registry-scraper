package scraper

import (
	"context"
	"sync"
	"time"

	"github.com/c-wide/albion-registry-scraper/internal/fetcher"
)

const PAGE_SIZE = 50
const MAX_OFFSET = 1000
const REQUEST_DELAY_DURATION = 150 * time.Millisecond

func (s *Scraper) PerformRecentEventCycle(ctx context.Context) {
	s.logger.Info().Msg("Starting recent event cycle...")
	cycleStart := time.Now()

	var wg sync.WaitGroup

	for region, f := range s.fetchers {
		wg.Add(1)

		go func(region string, f *fetcher.Fetcher) {
			defer wg.Done()

			s.logger.Info().Str("region", region).Msg("Fetching events...")
			fetchStart := time.Now()

			events, err := fetchAllRecentEvents(ctx, f, PAGE_SIZE, MAX_OFFSET)
			if err != nil {
				s.logger.Error().Err(err).Str("region", region).Msg("An error occurred while fetching events")
			}
			s.logger.Info().Str("region", region).Float64("fetchDuration", time.Since(fetchStart).Seconds()).Msg("Finished fetching events")

			if len(events) == 0 {
				s.logger.Info().Str("region", region).Msg("No events to process")
				return
			}

			uniqueEvents := make([]fetcher.KillboardEvent, 0, len(events))
			for _, event := range events {
				if _, ok := s.eventCaches[region].Get(event.EventID); ok {
					continue
				}
				s.eventCaches[region].Add(event.EventID, nil)
				uniqueEvents = append(uniqueEvents, event)
			}
			s.logger.Info().Str("region", region).Int("uniqueEventCount", len(uniqueEvents)).Msg("Finished extracting unique events")

			uniqueRatio := float64(len(uniqueEvents)) / float64(len(events))
			s.logger.Info().Str("region", region).Float64("uniqueRatio", uniqueRatio).Send()

			eventData := s.extractEventData(region, uniqueEvents)

			eCounts, err := s.persistEntities(ctx, region, eventData)
			if err != nil {
				s.logger.Error().Err(err).Str("region", region).Msg("Unable to create new entities")
				return
			}

			s.logger.Info().Str("region", region).Interface("entityCounts", eCounts).Msg("Successfully persisted entities")

			mCounts, err := s.persistMemberships(ctx, region, eventData)
			if err != nil {
				s.logger.Error().Err(err).Str("region", region).Msg("Failed to process player memberships")
			}

			s.logger.Info().Str("region", region).Interface("membershipCounts", mCounts).Msg("Successfully persisted memberships")
		}(region, f)
	}

	wg.Wait()

	s.logger.Info().Float64("eventCycleDuration", time.Since(cycleStart).Seconds()).Msg("Recent event cycle finished")
}

func fetchAllRecentEvents(ctx context.Context, f *fetcher.Fetcher, pageSize, maxOffset int) ([]fetcher.KillboardEvent, error) {
	var allEvents []fetcher.KillboardEvent
	for currentOffset := maxOffset; currentOffset >= 0; currentOffset -= pageSize {
		events, err := f.FetchRecentEvents(ctx, pageSize, currentOffset)
		if err != nil {
			return allEvents, err
		}
		allEvents = append(events, allEvents...)
		time.Sleep(REQUEST_DELAY_DURATION)
	}
	return allEvents, nil
}
