package scraper

import (
	"time"

	"github.com/c-wide/albion-registry-scraper/internal/database"
	"github.com/c-wide/albion-registry-scraper/internal/fetcher"
)

func newEventData() *EventData {
	return &EventData{
		Players:   make(map[string]*EventDataPlayer),
		Guilds:    make(map[string]*EventDataGuild),
		Alliances: make(map[string]*database.Alliance),
	}
}

func extractPlayers(event fetcher.KillboardEvent) []fetcher.Player {
	players := make([]fetcher.Player, 0, 2+len(event.Participants)+len(event.GroupMembers))
	players = append(players, event.Killer, event.Victim)
	players = append(players, event.Participants...)
	players = append(players, event.GroupMembers...)

	return players
}

func (s *Scraper) extractEventData(region string, events []fetcher.KillboardEvent) *EventData {
	eventData := newEventData()

	for _, event := range events {
		timestamp, err := time.Parse(time.RFC3339Nano, event.Timestamp)
		if err != nil {
			s.logger.Error().Err(err).Str("region", region).Uint64("eventId", event.EventID).Msg("Error parsing timestamp for event")
			continue
		}

		timestamp = timestamp.Round(time.Microsecond)

		players := extractPlayers(event)
		for _, player := range players {
			// Check if player already exists in this event batch
			_, ok := eventData.Players[player.ID]

			// If the player doesn't exist, create them
			// otherwise update the FirstSeen timestamp
			if !ok {
				eventData.Players[player.ID] = &EventDataPlayer{
					Info: database.Player{
						Name:       player.Name,
						PlayerID:   player.ID,
						Region:     region,
						Avatar:     &player.Avatar,
						AvatarRing: &player.AvatarRing,
						FirstSeen:  timestamp,
						LastSeen:   timestamp,
					},
				}
			} else {
				eventData.Players[player.ID].Info.FirstSeen = timestamp
			}

			// If player was in a guild for this event, process data
			if player.GuildID != "" {
				// Check if guild already exists in this event batch
				_, ok := eventData.Guilds[player.GuildID]

				// If the guild doesn't exist, create it
				// otherwise update the FirstSeen timestamp
				if !ok {
					eventData.Guilds[player.GuildID] = &EventDataGuild{
						Info: database.Guild{
							Name:      player.GuildName,
							GuildID:   player.GuildID,
							Region:    region,
							FirstSeen: timestamp,
							LastSeen:  timestamp,
						},
					}
				} else {
					eventData.Guilds[player.GuildID].Info.FirstSeen = timestamp
				}

				// Process the player guild history
				//
				// If the player has no guild history in this event batch,
				// skip processing and append the data
				if len(eventData.Players[player.ID].GuildHistory) == 0 {
					eventData.Players[player.ID].GuildHistory = append(eventData.Players[player.ID].GuildHistory, database.CreatePGMsParams{
						PlayerID:  player.ID,
						GuildID:   player.GuildID,
						Region:    region,
						IsActive:  true,
						FirstSeen: timestamp,
						LastSeen:  timestamp,
					})
				} else {
					oldestGuild := &eventData.Players[player.ID].GuildHistory[len(eventData.Players[player.ID].GuildHistory)-1]

					// If the player's *oldest* guild is the same as this event's current guild
					// update the FirstSeen timestamp, otherwise, append the guild membership data
					if oldestGuild.GuildID == player.GuildID {
						oldestGuild.FirstSeen = timestamp
					} else {
						eventData.Players[player.ID].GuildHistory = append(eventData.Players[player.ID].GuildHistory, database.CreatePGMsParams{
							PlayerID:  player.ID,
							GuildID:   player.GuildID,
							Region:    region,
							IsActive:  false,
							FirstSeen: timestamp,
							LastSeen:  timestamp,
						})
					}
				}
			}

			// If player was in an alliance for this event, process data
			if player.AllianceID != "" {
				// Check if alliance already exists in this event batch
				_, ok := eventData.Alliances[player.AllianceID]

				// If the alliance doesn't exist, create it
				// otherwise update the FirstSeen timestamp
				if !ok {
					eventData.Alliances[player.AllianceID] = &database.Alliance{
						Tag:        player.AllianceTag,
						AllianceID: player.AllianceID,
						Region:     region,
						FirstSeen:  timestamp,
						LastSeen:   timestamp,
					}
				} else {
					eventData.Alliances[player.AllianceID].FirstSeen = timestamp
				}

				// Process the guild alliance history
				//
				// If the guild has no alliance history in this event batch,
				// skip processing and append the data
				if len(eventData.Guilds[player.GuildID].AllianceHistory) == 0 {
					eventData.Guilds[player.GuildID].AllianceHistory = append(eventData.Guilds[player.GuildID].AllianceHistory, database.CreateGAMsParams{
						GuildID:    player.GuildID,
						AllianceID: player.AllianceID,
						Region:     region,
						IsActive:   true,
						FirstSeen:  timestamp,
						LastSeen:   timestamp,
					})
				} else {
					oldestAlliance := &eventData.Guilds[player.GuildID].AllianceHistory[len(eventData.Guilds[player.GuildID].AllianceHistory)-1]

					// If the guild's **oldest** alliance is the same as this event's current alliance
					// update the FirstSeen timestamp, otherwise, append the alliance membership data
					if oldestAlliance.AllianceID == player.AllianceID {
						oldestAlliance.FirstSeen = timestamp
					} else {
						eventData.Guilds[player.GuildID].AllianceHistory = append(eventData.Guilds[player.GuildID].AllianceHistory, database.CreateGAMsParams{
							GuildID:    player.GuildID,
							AllianceID: player.AllianceID,
							Region:     region,
							IsActive:   false,
							FirstSeen:  timestamp,
							LastSeen:   timestamp,
						})
					}
				}
			}
		}
	}

	return eventData
}
