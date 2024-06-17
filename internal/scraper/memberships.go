package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/c-wide/albion-registry-scraper/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Scraper) persistMemberships(ctx context.Context, region string, eventData *EventData) (*MembershipResults, error) {
	pgmCounts, err := persistPGMs(ctx, s.queries, region, eventData.Players)
	if err != nil {
		return nil, err
	}

	gamCounts, err := persistGAMs(ctx, s.queries, region, eventData.Guilds)
	if err != nil {
		return nil, err
	}

	counts := MembershipResults{
		PGM: pgmCounts,
		GAM: gamCounts,
	}

	return &counts, nil
}

func persistPGMs(ctx context.Context, queries *database.Queries, region string, players map[string]*EventDataPlayer) (*ChangeCounts, error) {
	playerIds := make([]string, 0, len(players))

	for id := range players {
		playerIds = append(playerIds, id)
	}

	memberships, err := queries.GetLatestPGMs(ctx, database.GetLatestPGMsParams{
		Ids:    playerIds,
		Region: region,
	})

	if err != nil {
		return nil, fmt.Errorf("unable to query latest player guild memberships: %w", err)
	}

	var membershipMap = make(map[string]database.PlayerGuildMembership, len(memberships))
	for _, membership := range memberships {
		membershipMap[membership.PlayerID] = membership
	}

	var inactiveRecordIDs []pgtype.UUID
	var historyRecordIDs []pgtype.UUID
	var newTimestamps []time.Time
	var newMemberships []database.CreatePGMsParams

	for _, player := range players {
		membership, ok := membershipMap[player.Info.PlayerID]

		if !ok {
			if len(player.GuildHistory) == 0 {
				continue
			}

			newMemberships = append(newMemberships, player.GuildHistory...)
			continue
		}

		if len(player.GuildHistory) == 0 {
			inactiveRecordIDs = append(inactiveRecordIDs, membership.ID)
			continue
		}

		oldestGuild := player.GuildHistory[len(player.GuildHistory)-1]
		if membership.GuildID == oldestGuild.GuildID {
			historyRecordIDs = append(historyRecordIDs, membership.ID)
			newTimestamps = append(newTimestamps, oldestGuild.LastSeen)

			if len(player.GuildHistory) > 1 {
				inactiveRecordIDs = append(inactiveRecordIDs, membership.ID)
				newMemberships = append(newMemberships, player.GuildHistory[:len(player.GuildHistory)-1]...)
			}
		} else {
			inactiveRecordIDs = append(inactiveRecordIDs, membership.ID)
			newMemberships = append(newMemberships, player.GuildHistory...)
		}
	}

	if len(historyRecordIDs) > 0 {
		err = queries.UpdatePGMsLastSeen(ctx, database.UpdatePGMsLastSeenParams{
			Ids:        historyRecordIDs,
			Timestamps: newTimestamps,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to update player guild memberships: %w", err)
		}
	}

	if len(inactiveRecordIDs) > 0 {
		err = queries.SetPGMsInactive(ctx, inactiveRecordIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to set player guild memberships inactive: %w", err)
		}
	}

	if len(newMemberships) > 0 {
		_, err = queries.CreatePGMs(ctx, newMemberships)
		if err != nil {
			return nil, fmt.Errorf("failed to create player guild memberships: %w", err)
		}
	}

	changeCounts := ChangeCounts{
		New:     len(newMemberships),
		Updated: len(historyRecordIDs) + len(inactiveRecordIDs),
	}

	return &changeCounts, nil
}

func persistGAMs(ctx context.Context, queries *database.Queries, region string, guilds map[string]*EventDataGuild) (*ChangeCounts, error) {
	guildIds := make([]string, 0, len(guilds))

	for id := range guilds {
		guildIds = append(guildIds, id)
	}

	memberships, err := queries.GetLatestGAMs(ctx, database.GetLatestGAMsParams{
		Ids:    guildIds,
		Region: region,
	})

	if err != nil {
		return nil, fmt.Errorf("unable to query latest guild alliance memberships: %w", err)
	}

	var membershipMap = make(map[string]database.GuildAllianceMembership, len(memberships))
	for _, membership := range memberships {
		membershipMap[membership.GuildID] = membership
	}

	var inactiveRecordIDs []pgtype.UUID
	var historyRecordIDs []pgtype.UUID
	var newTimestamps []time.Time
	var newMemberships []database.CreateGAMsParams

	for _, guild := range guilds {
		membership, ok := membershipMap[guild.Info.GuildID]

		if !ok {
			if len(guild.AllianceHistory) == 0 {
				continue
			}

			newMemberships = append(newMemberships, guild.AllianceHistory...)
			continue
		}

		if len(guild.AllianceHistory) == 0 {
			inactiveRecordIDs = append(inactiveRecordIDs, membership.ID)
			continue
		}

		oldestAlliance := guild.AllianceHistory[len(guild.AllianceHistory)-1]
		if membership.AllianceID == oldestAlliance.AllianceID {
			historyRecordIDs = append(historyRecordIDs, membership.ID)
			newTimestamps = append(newTimestamps, oldestAlliance.LastSeen)

			if len(guild.AllianceHistory) > 1 {
				inactiveRecordIDs = append(inactiveRecordIDs, membership.ID)
				newMemberships = append(newMemberships, guild.AllianceHistory[:len(guild.AllianceHistory)-1]...)
			}
		} else {
			inactiveRecordIDs = append(inactiveRecordIDs, membership.ID)
			newMemberships = append(newMemberships, guild.AllianceHistory...)
		}
	}

	if len(historyRecordIDs) > 0 {
		err = queries.UpdateGAMsLastSeen(ctx, database.UpdateGAMsLastSeenParams{
			Ids:        historyRecordIDs,
			Timestamps: newTimestamps,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to update guild alliance memberships: %w", err)
		}
	}

	if len(inactiveRecordIDs) > 0 {
		err = queries.SetGAMsInactive(ctx, inactiveRecordIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to set guild alliance memberships inactive: %w", err)
		}
	}

	if len(newMemberships) > 0 {
		_, err = queries.CreateGAMs(ctx, newMemberships)
		if err != nil {
			return nil, fmt.Errorf("failed to create guild alliance memberships: %w", err)
		}
	}

	changeCounts := ChangeCounts{
		New:     len(newMemberships),
		Updated: len(historyRecordIDs) + len(inactiveRecordIDs),
	}

	return &changeCounts, nil
}
