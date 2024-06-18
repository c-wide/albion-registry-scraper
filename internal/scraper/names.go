package scraper

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/c-wide/albion-registry-scraper/internal/database"
	"github.com/c-wide/albion-registry-scraper/internal/fetcher"
	"github.com/jackc/pgx/v5"
)

func (s *Scraper) PerformAllianceNameCycle(ctx context.Context) {
	s.logger.Info().Msg("Starting alliance name cycle")

	cycleStart := time.Now()

	var wg sync.WaitGroup

	for r, f := range s.fetchers {
		wg.Add(1)

		go func(region string, fetcher *fetcher.Fetcher) {
			defer wg.Done()

			allianceID, err := s.queries.GetNullNameAlliance(ctx, region)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return
				}

				s.logger.Error().Err(err).Str("region", region).Msg("Failed to query database for nameless alliance")
				return
			}

			allianceInfo, err := fetcher.FetchAllianceInfo(ctx, allianceID)
			if err != nil {
				s.logger.Error().Err(err).Str("region", region).Msg("Failed to fetch alliance information from Albion Online API")

				if strings.Contains(err.Error(), "404") {
					err = s.queries.SetAllianceSkipName(ctx, database.SetAllianceSkipNameParams{
						ID:     allianceID,
						Region: region,
					})

					if err != nil {
						s.logger.Error().Err(err).Str("region", region).Str("allianceId", allianceID).Msg("Failed to set alliance skip name")
						return
					}

					s.logger.Info().Str("region", region).Str("allianceId", allianceID).Msg("Set alliance skip name successfully")
				}

				return
			}

			err = s.queries.SetAllianceName(ctx, database.SetAllianceNameParams{
				Name:   &allianceInfo.Name,
				ID:     allianceID,
				Region: region,
			})

			if err != nil {
				s.logger.Error().Err(err).Str("region", region).Msg("Failed to set alliance name")
				return
			}
		}(r, f)
	}

	wg.Wait()

	s.logger.Info().Float64("nameCycleDuration", time.Since(cycleStart).Seconds()).Msg("Finished fetching alliance names")
}
