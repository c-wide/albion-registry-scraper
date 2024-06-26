// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: copyfrom.go

package database

import (
	"context"
)

// iteratorForCreateGAMs implements pgx.CopyFromSource.
type iteratorForCreateGAMs struct {
	rows                 []CreateGAMsParams
	skippedFirstNextCall bool
}

func (r *iteratorForCreateGAMs) Next() bool {
	if len(r.rows) == 0 {
		return false
	}
	if !r.skippedFirstNextCall {
		r.skippedFirstNextCall = true
		return true
	}
	r.rows = r.rows[1:]
	return len(r.rows) > 0
}

func (r iteratorForCreateGAMs) Values() ([]interface{}, error) {
	return []interface{}{
		r.rows[0].GuildID,
		r.rows[0].AllianceID,
		r.rows[0].Region,
		r.rows[0].IsActive,
		r.rows[0].FirstSeen,
		r.rows[0].LastSeen,
	}, nil
}

func (r iteratorForCreateGAMs) Err() error {
	return nil
}

func (q *Queries) CreateGAMs(ctx context.Context, arg []CreateGAMsParams) (int64, error) {
	return q.db.CopyFrom(ctx, []string{"guild_alliance_memberships"}, []string{"guild_id", "alliance_id", "region", "is_active", "first_seen", "last_seen"}, &iteratorForCreateGAMs{rows: arg})
}

// iteratorForCreatePGMs implements pgx.CopyFromSource.
type iteratorForCreatePGMs struct {
	rows                 []CreatePGMsParams
	skippedFirstNextCall bool
}

func (r *iteratorForCreatePGMs) Next() bool {
	if len(r.rows) == 0 {
		return false
	}
	if !r.skippedFirstNextCall {
		r.skippedFirstNextCall = true
		return true
	}
	r.rows = r.rows[1:]
	return len(r.rows) > 0
}

func (r iteratorForCreatePGMs) Values() ([]interface{}, error) {
	return []interface{}{
		r.rows[0].PlayerID,
		r.rows[0].GuildID,
		r.rows[0].Region,
		r.rows[0].IsActive,
		r.rows[0].FirstSeen,
		r.rows[0].LastSeen,
	}, nil
}

func (r iteratorForCreatePGMs) Err() error {
	return nil
}

func (q *Queries) CreatePGMs(ctx context.Context, arg []CreatePGMsParams) (int64, error) {
	return q.db.CopyFrom(ctx, []string{"player_guild_memberships"}, []string{"player_id", "guild_id", "region", "is_active", "first_seen", "last_seen"}, &iteratorForCreatePGMs{rows: arg})
}
