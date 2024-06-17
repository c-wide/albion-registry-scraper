package scraper

import "github.com/c-wide/albion-registry-scraper/internal/database"

type EventDataPlayer struct {
	Info         database.Player
	GuildHistory []database.CreatePGMsParams
}

type EventDataGuild struct {
	Info            database.Guild
	AllianceHistory []database.CreateGAMsParams
}

type EventData struct {
	Players   map[string]*EventDataPlayer
	Guilds    map[string]*EventDataGuild
	Alliances map[string]*database.Alliance
}

type ChangeCounts struct {
	New     int
	Updated int
}

type UpsertResults struct {
	Players   *ChangeCounts
	Guilds    *ChangeCounts
	Alliances *ChangeCounts
}

type MembershipResults struct {
	PGM *ChangeCounts
	GAM *ChangeCounts
}
