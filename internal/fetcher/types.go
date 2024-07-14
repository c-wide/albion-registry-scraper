package fetcher

import (
	"net/http"
)

type Fetcher struct {
	httpClient *http.Client
	baseURL    string
}

type Player struct {
	Name        string `json:"Name"`
	ID          string `json:"Id"`
	GuildName   string `json:"GuildName"`
	GuildID     string `json:"GuildId"`
	AllianceTag string `json:"AllianceName"`
	AllianceID  string `json:"AllianceId"`
	Avatar      string `json:"Avatar"`
	AvatarRing  string `json:"AvatarRing"`
}

type KillboardEvent struct {
	EventID      uint64   `json:"EventId"`
	Timestamp    string   `json:"TimeStamp"`
	Killer       Player   `json:"Killer"`
	Victim       Player   `json:"Victim"`
	Participants []Player `json:"Participants"`
	GroupMembers []Player `json:"GroupMembers"`
}

type Alliance struct {
	Name string `json:"AllianceName"`
	Tag  string `json:"AllianceTag"`
}
