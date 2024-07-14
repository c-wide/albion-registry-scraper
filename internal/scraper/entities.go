package scraper

import (
	"context"
	"time"

	"github.com/c-wide/albion-registry-scraper/internal/database"
)

func (s *Scraper) persistEntities(ctx context.Context, region string, eventData *EventData) (*UpsertResults, error) {
	pCounts, err := upsertPlayers(ctx, s.queries, eventData)
	if err != nil {
		s.logger.Error().Err(err).Str("region", region).Msg("Error upserting players")
	}

	gCounts, err := upsertGuilds(ctx, s.queries, eventData)
	if err != nil {
		s.logger.Error().Err(err).Str("region", region).Msg("Error upserting guilds")
	}

	aCounts, err := upsertAlliances(ctx, s.queries, eventData)
	if err != nil {
		s.logger.Error().Err(err).Str("region", region).Msg("Error upserting alliances")
	}

	upsertResults := UpsertResults{
		Players:   pCounts,
		Guilds:    gCounts,
		Alliances: aCounts,
	}

	return &upsertResults, nil
}

func upsertPlayers(ctx context.Context, queries *database.Queries, eventData *EventData) (*ChangeCounts, error) {
	pCount := len(eventData.Players)

	ids := make([]string, 0, pCount)
	names := make([]string, 0, pCount)
	regions := make([]string, 0, pCount)
	avatars := make([]string, 0, pCount)
	avatarRings := make([]string, 0, pCount)
	fsts := make([]time.Time, 0, pCount)
	lsts := make([]time.Time, 0, pCount)

	for _, player := range eventData.Players {
		ids = append(ids, player.Info.PlayerID)
		names = append(names, player.Info.Name)
		regions = append(regions, player.Info.Region)
		avatars = append(avatars, *player.Info.Avatar)
		avatarRings = append(avatarRings, *player.Info.AvatarRing)
		fsts = append(fsts, player.Info.FirstSeen)
		lsts = append(lsts, player.Info.LastSeen)
	}

	rows, err := queries.UpsertPlayers(ctx, database.UpsertPlayersParams{
		Ids:         ids,
		Names:       names,
		Regions:     regions,
		Avatars:     avatars,
		AvatarRings: avatarRings,
		Fsts:        fsts,
		Lsts:        lsts,
	})

	if err != nil {
		return nil, err
	}

	upsertResult := ChangeCounts{
		New:     0,
		Updated: 0,
	}

	for _, row := range rows {
		if time.Time.Equal(row.FirstSeen.UTC(), eventData.Players[row.PlayerID].Info.FirstSeen) {
			upsertResult.New++
		} else {
			upsertResult.Updated++
		}
	}

	return &upsertResult, nil
}

func upsertGuilds(ctx context.Context, queries *database.Queries, eventData *EventData) (*ChangeCounts, error) {
	gCount := len(eventData.Guilds)

	ids := make([]string, 0, gCount)
	names := make([]string, 0, gCount)
	regions := make([]string, 0, gCount)
	fsts := make([]time.Time, 0, gCount)
	lsts := make([]time.Time, 0, gCount)

	for _, guild := range eventData.Guilds {
		ids = append(ids, guild.Info.GuildID)
		names = append(names, guild.Info.Name)
		regions = append(regions, guild.Info.Region)
		fsts = append(fsts, guild.Info.FirstSeen)
		lsts = append(lsts, guild.Info.LastSeen)
	}

	rows, err := queries.UpsertGuilds(ctx, database.UpsertGuildsParams{
		Ids:     ids,
		Names:   names,
		Regions: regions,
		Fsts:    fsts,
		Lsts:    lsts,
	})

	if err != nil {
		return nil, err
	}

	upsertResult := ChangeCounts{
		New:     0,
		Updated: 0,
	}

	for _, row := range rows {
		if time.Time.Equal(row.FirstSeen.UTC(), eventData.Guilds[row.GuildID].Info.FirstSeen) {
			upsertResult.New++
		} else {
			upsertResult.Updated++
		}
	}

	return &upsertResult, nil
}

func upsertAlliances(ctx context.Context, queries *database.Queries, eventData *EventData) (*ChangeCounts, error) {
	aCount := len(eventData.Alliances)

	ids := make([]string, 0, aCount)
	tags := make([]string, 0, aCount)
	regions := make([]string, 0, aCount)
	fsts := make([]time.Time, 0, aCount)
	lsts := make([]time.Time, 0, aCount)

	for _, alliance := range eventData.Alliances {
		ids = append(ids, alliance.AllianceID)
		tags = append(tags, alliance.Tag)
		regions = append(regions, alliance.Region)
		fsts = append(fsts, alliance.FirstSeen)
		lsts = append(lsts, alliance.LastSeen)
	}

	rows, err := queries.UpsertAlliances(ctx, database.UpsertAlliancesParams{
		Ids:     ids,
		Tags:    tags,
		Regions: regions,
		Fsts:    fsts,
		Lsts:    lsts,
	})

	if err != nil {
		return nil, err
	}

	upsertResult := ChangeCounts{
		New:     0,
		Updated: 0,
	}

	for _, row := range rows {
		if time.Time.Equal(row.FirstSeen.UTC(), eventData.Alliances[row.AllianceID].FirstSeen) {
			upsertResult.New++
		} else {
			upsertResult.Updated++
		}
	}

	return &upsertResult, nil
}
