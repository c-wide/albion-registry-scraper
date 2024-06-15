-- name: UpsertPlayers :many
INSERT INTO players (player_id, name, region, first_seen, last_seen)
VALUES (
    unnest(@ids::varchar(50)[]),
    unnest(@names::varchar(25)[]),
    unnest(@regions::varchar(50)[]),
    unnest(@fsts::timestamptz[]),
    unnest(@lsts::timestamptz[])
)
ON CONFLICT (player_id, region) DO UPDATE
SET last_seen = EXCLUDED.last_seen
RETURNING player_id, first_seen;

-- name: UpsertGuilds :many
INSERT INTO guilds (guild_id, name, region, first_seen, last_seen)
VALUES (
    unnest(@ids::varchar(50)[]),
    unnest(@names::varchar(50)[]),
    unnest(@regions::varchar(50)[]),
    unnest(@fsts::timestamptz[]),
    unnest(@lsts::timestamptz[])
)
ON CONFLICT (guild_id, region) DO UPDATE
SET last_seen = EXCLUDED.last_seen
RETURNING guild_id, first_seen;

-- name: UpsertAlliances :many
INSERT INTO alliances (alliance_id, tag, region, first_seen, last_seen)
VALUES (
    unnest(@ids::varchar(50)[]),
    unnest(@tags::varchar(5)[]),
    unnest(@regions::varchar(50)[]),
    unnest(@fsts::timestamptz[]),
    unnest(@lsts::timestamptz[])
)
ON CONFLICT (alliance_id, region) DO UPDATE
SET last_seen = EXCLUDED.last_seen
RETURNING alliance_id, first_seen;

-- name: GetLatestPGMs :many
SELECT pgm.*
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
            p.player_id = ANY(@ids::varchar(50)[])
            AND p.region = @region
            AND p.is_active = true
        GROUP BY
            p.player_id
    ) latest_pgm ON pgm.player_id = latest_pgm.player_id
                 AND pgm.last_seen = latest_pgm.latest_last_seen
WHERE
    pgm.player_id = ANY(@ids::varchar(50)[])
    AND pgm.region = @region
    AND pgm.is_active = true;

-- name: GetLatestGAMs :many
SELECT gam.*
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
            g.guild_id = ANY(@ids::varchar(50)[])
            AND g.region = @region
            AND g.is_active = true
        GROUP BY
            g.guild_id
    ) latest_gam ON gam.guild_id = latest_gam.guild_id
                 AND gam.last_seen = latest_gam.latest_last_seen
WHERE
    gam.guild_id = ANY(@ids::varchar(50)[])
    AND gam.region = @region
    AND gam.is_active = true;

-- name: UpdatePGMsLastSeen :exec
UPDATE player_guild_memberships AS pgm
SET last_seen = d.new_last_seen
FROM
    (SELECT
        unnest(@ids::uuid[]) AS record_id,
        unnest(@timestamps::timestamptz[]) AS new_last_seen
    ) AS d
WHERE pgm.id = d.record_id;

-- name: UpdateGAMsLastSeen :exec
UPDATE guild_alliance_memberships AS gam
SET last_seen = d.new_last_seen
FROM
    (SELECT
        unnest(@ids::uuid[]) AS record_id,
        unnest(@timestamps::timestamptz[]) AS new_last_seen
    ) AS d
WHERE gam.id = d.record_id;

-- name: CreatePGMs :copyfrom
INSERT INTO player_guild_memberships (player_id, guild_id, region, is_active, first_seen, last_seen)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: CreateGAMs :copyfrom
INSERT INTO guild_alliance_memberships (guild_id, alliance_id, region, is_active, first_seen, last_seen)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: SetPGMsInactive :exec
UPDATE player_guild_memberships AS pgm
SET is_active = false
FROM
    (SELECT
        unnest(@ids::uuid[]) as record_id
    ) as d
WHERE pgm.id = d.record_id;

-- name: SetGAMsInactive :exec
UPDATE guild_alliance_memberships AS gam
SET is_active = false
FROM
    (SELECT
        unnest(@ids::uuid[]) as record_id
    ) as d
WHERE gam.id = d.record_id;

-- name: GetNullNameAlliance :one
SELECT alliance_id
FROM alliances
WHERE region = @region
AND name IS NULL
AND skip_name_check = false
ORDER BY first_seen
LIMIT 1;

-- name: SetAllianceSkipName :exec
UPDATE alliances
SET skip_name_check = true
WHERE alliance_id = @id
AND region = @region;

-- name: SetAllianceName :exec
UPDATE alliances
SET name = @name
WHERE alliance_id = @id
AND region = @region;
