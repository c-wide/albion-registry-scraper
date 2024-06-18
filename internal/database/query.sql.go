// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: query.sql

package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreateGAMsParams struct {
	GuildID    string
	AllianceID string
	Region     string
	IsActive   bool
	FirstSeen  time.Time
	LastSeen   time.Time
}

type CreatePGMsParams struct {
	PlayerID  string
	GuildID   string
	Region    string
	IsActive  bool
	FirstSeen time.Time
	LastSeen  time.Time
}

const getLatestGAMs = `-- name: GetLatestGAMs :many
SELECT gam.id, gam.guild_id, gam.alliance_id, gam.region, gam.is_active, gam.first_seen, gam.last_seen
FROM
    guild_alliance_memberships gam
JOIN
    (
        SELECT
            g.guild_id,
            MAX(g.last_seen) AS latest_last_seen
        FROM
            guild_alliance_memberships g
        WHERE
            g.guild_id = ANY($1::varchar(50)[])
            AND g.region = $2
            AND g.is_active = true
        GROUP BY
            g.guild_id
    ) latest_gam ON gam.guild_id = latest_gam.guild_id
                 AND gam.last_seen = latest_gam.latest_last_seen
WHERE
    gam.guild_id = ANY($1::varchar(50)[])
    AND gam.region = $2
    AND gam.is_active = true
`

type GetLatestGAMsParams struct {
	Ids    []string
	Region string
}

func (q *Queries) GetLatestGAMs(ctx context.Context, arg GetLatestGAMsParams) ([]GuildAllianceMembership, error) {
	rows, err := q.db.Query(ctx, getLatestGAMs, arg.Ids, arg.Region)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GuildAllianceMembership{}
	for rows.Next() {
		var i GuildAllianceMembership
		if err := rows.Scan(
			&i.ID,
			&i.GuildID,
			&i.AllianceID,
			&i.Region,
			&i.IsActive,
			&i.FirstSeen,
			&i.LastSeen,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getLatestPGMs = `-- name: GetLatestPGMs :many
SELECT pgm.id, pgm.player_id, pgm.guild_id, pgm.region, pgm.is_active, pgm.first_seen, pgm.last_seen
FROM
    player_guild_memberships pgm
JOIN
    (
        SELECT
            p.player_id,
            MAX(p.last_seen) AS latest_last_seen
        FROM
            player_guild_memberships p
        WHERE
            p.player_id = ANY($1::varchar(50)[])
            AND p.region = $2
            AND p.is_active = true
        GROUP BY
            p.player_id
    ) latest_pgm ON pgm.player_id = latest_pgm.player_id
                 AND pgm.last_seen = latest_pgm.latest_last_seen
WHERE
    pgm.player_id = ANY($1::varchar(50)[])
    AND pgm.region = $2
    AND pgm.is_active = true
`

type GetLatestPGMsParams struct {
	Ids    []string
	Region string
}

func (q *Queries) GetLatestPGMs(ctx context.Context, arg GetLatestPGMsParams) ([]PlayerGuildMembership, error) {
	rows, err := q.db.Query(ctx, getLatestPGMs, arg.Ids, arg.Region)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []PlayerGuildMembership{}
	for rows.Next() {
		var i PlayerGuildMembership
		if err := rows.Scan(
			&i.ID,
			&i.PlayerID,
			&i.GuildID,
			&i.Region,
			&i.IsActive,
			&i.FirstSeen,
			&i.LastSeen,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getNullNameAlliance = `-- name: GetNullNameAlliance :one
SELECT alliance_id
FROM alliances
WHERE region = $1
AND name IS NULL
AND skip_name_check = false
ORDER BY first_seen
LIMIT 1
`

func (q *Queries) GetNullNameAlliance(ctx context.Context, region string) (string, error) {
	row := q.db.QueryRow(ctx, getNullNameAlliance, region)
	var alliance_id string
	err := row.Scan(&alliance_id)
	return alliance_id, err
}

const setAllianceName = `-- name: SetAllianceName :exec
UPDATE alliances
SET name = $1
WHERE alliance_id = $2
AND region = $3
`

type SetAllianceNameParams struct {
	Name   *string
	ID     string
	Region string
}

func (q *Queries) SetAllianceName(ctx context.Context, arg SetAllianceNameParams) error {
	_, err := q.db.Exec(ctx, setAllianceName, arg.Name, arg.ID, arg.Region)
	return err
}

const setAllianceSkipName = `-- name: SetAllianceSkipName :exec
UPDATE alliances
SET skip_name_check = true
WHERE alliance_id = $1
AND region = $2
AND skip_name_check = false
`

type SetAllianceSkipNameParams struct {
	ID     string
	Region string
}

func (q *Queries) SetAllianceSkipName(ctx context.Context, arg SetAllianceSkipNameParams) error {
	_, err := q.db.Exec(ctx, setAllianceSkipName, arg.ID, arg.Region)
	return err
}

const setGAMsInactive = `-- name: SetGAMsInactive :exec
UPDATE guild_alliance_memberships AS gam
SET is_active = false
FROM
    (SELECT
        unnest($1::uuid[]) as record_id
    ) as d
WHERE gam.id = d.record_id
`

func (q *Queries) SetGAMsInactive(ctx context.Context, ids []pgtype.UUID) error {
	_, err := q.db.Exec(ctx, setGAMsInactive, ids)
	return err
}

const setPGMsInactive = `-- name: SetPGMsInactive :exec
UPDATE player_guild_memberships AS pgm
SET is_active = false
FROM
    (SELECT
        unnest($1::uuid[]) as record_id
    ) as d
WHERE pgm.id = d.record_id
`

func (q *Queries) SetPGMsInactive(ctx context.Context, ids []pgtype.UUID) error {
	_, err := q.db.Exec(ctx, setPGMsInactive, ids)
	return err
}

const updateGAMsLastSeen = `-- name: UpdateGAMsLastSeen :exec
UPDATE guild_alliance_memberships AS gam
SET last_seen = d.new_last_seen
FROM
    (SELECT
        unnest($1::uuid[]) AS record_id,
        unnest($2::timestamptz[]) AS new_last_seen
    ) AS d
WHERE gam.id = d.record_id
`

type UpdateGAMsLastSeenParams struct {
	Ids        []pgtype.UUID
	Timestamps []time.Time
}

func (q *Queries) UpdateGAMsLastSeen(ctx context.Context, arg UpdateGAMsLastSeenParams) error {
	_, err := q.db.Exec(ctx, updateGAMsLastSeen, arg.Ids, arg.Timestamps)
	return err
}

const updatePGMsLastSeen = `-- name: UpdatePGMsLastSeen :exec
UPDATE player_guild_memberships AS pgm
SET last_seen = d.new_last_seen
FROM
    (SELECT
        unnest($1::uuid[]) AS record_id,
        unnest($2::timestamptz[]) AS new_last_seen
    ) AS d
WHERE pgm.id = d.record_id
`

type UpdatePGMsLastSeenParams struct {
	Ids        []pgtype.UUID
	Timestamps []time.Time
}

func (q *Queries) UpdatePGMsLastSeen(ctx context.Context, arg UpdatePGMsLastSeenParams) error {
	_, err := q.db.Exec(ctx, updatePGMsLastSeen, arg.Ids, arg.Timestamps)
	return err
}

const upsertAlliances = `-- name: UpsertAlliances :many
INSERT INTO alliances (alliance_id, tag, region, first_seen, last_seen)
VALUES (
    unnest($1::varchar(50)[]),
    unnest($2::varchar(5)[]),
    unnest($3::varchar(50)[]),
    unnest($4::timestamptz[]),
    unnest($5::timestamptz[])
)
ON CONFLICT (alliance_id, region) DO UPDATE
SET last_seen = EXCLUDED.last_seen
RETURNING alliance_id, first_seen
`

type UpsertAlliancesParams struct {
	Ids     []string
	Tags    []string
	Regions []string
	Fsts    []time.Time
	Lsts    []time.Time
}

type UpsertAlliancesRow struct {
	AllianceID string
	FirstSeen  time.Time
}

func (q *Queries) UpsertAlliances(ctx context.Context, arg UpsertAlliancesParams) ([]UpsertAlliancesRow, error) {
	rows, err := q.db.Query(ctx, upsertAlliances,
		arg.Ids,
		arg.Tags,
		arg.Regions,
		arg.Fsts,
		arg.Lsts,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []UpsertAlliancesRow{}
	for rows.Next() {
		var i UpsertAlliancesRow
		if err := rows.Scan(&i.AllianceID, &i.FirstSeen); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const upsertGuilds = `-- name: UpsertGuilds :many
INSERT INTO guilds (guild_id, name, region, first_seen, last_seen)
VALUES (
    unnest($1::varchar(50)[]),
    unnest($2::varchar(50)[]),
    unnest($3::varchar(50)[]),
    unnest($4::timestamptz[]),
    unnest($5::timestamptz[])
)
ON CONFLICT (guild_id, region) DO UPDATE
SET last_seen = EXCLUDED.last_seen
RETURNING guild_id, first_seen
`

type UpsertGuildsParams struct {
	Ids     []string
	Names   []string
	Regions []string
	Fsts    []time.Time
	Lsts    []time.Time
}

type UpsertGuildsRow struct {
	GuildID   string
	FirstSeen time.Time
}

func (q *Queries) UpsertGuilds(ctx context.Context, arg UpsertGuildsParams) ([]UpsertGuildsRow, error) {
	rows, err := q.db.Query(ctx, upsertGuilds,
		arg.Ids,
		arg.Names,
		arg.Regions,
		arg.Fsts,
		arg.Lsts,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []UpsertGuildsRow{}
	for rows.Next() {
		var i UpsertGuildsRow
		if err := rows.Scan(&i.GuildID, &i.FirstSeen); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const upsertPlayers = `-- name: UpsertPlayers :many
INSERT INTO players (player_id, name, region, first_seen, last_seen)
VALUES (
    unnest($1::varchar(50)[]),
    unnest($2::varchar(25)[]),
    unnest($3::varchar(50)[]),
    unnest($4::timestamptz[]),
    unnest($5::timestamptz[])
)
ON CONFLICT (player_id, region) DO UPDATE
SET last_seen = EXCLUDED.last_seen
RETURNING player_id, first_seen
`

type UpsertPlayersParams struct {
	Ids     []string
	Names   []string
	Regions []string
	Fsts    []time.Time
	Lsts    []time.Time
}

type UpsertPlayersRow struct {
	PlayerID  string
	FirstSeen time.Time
}

func (q *Queries) UpsertPlayers(ctx context.Context, arg UpsertPlayersParams) ([]UpsertPlayersRow, error) {
	rows, err := q.db.Query(ctx, upsertPlayers,
		arg.Ids,
		arg.Names,
		arg.Regions,
		arg.Fsts,
		arg.Lsts,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []UpsertPlayersRow{}
	for rows.Next() {
		var i UpsertPlayersRow
		if err := rows.Scan(&i.PlayerID, &i.FirstSeen); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
