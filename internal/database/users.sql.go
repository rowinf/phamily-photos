// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: users.sql

package database

import (
	"context"
	"database/sql"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name, apikey, password, family_id)
VALUES ($1, NOW(), NOW(), $2, encode(sha256(random()::text::bytea), 'hex'), $3, $4)
RETURNING id, created_at, updated_at, name, apikey, family_id, password
`

type CreateUserParams struct {
	ID       string
	Name     string
	Password string
	FamilyID sql.NullInt64
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.ID,
		arg.Name,
		arg.Password,
		arg.FamilyID,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Apikey,
		&i.FamilyID,
		&i.Password,
	)
	return i, err
}

const getUserByApiKey = `-- name: GetUserByApiKey :one
SELECT id, created_at, updated_at, name, apikey, family_id, password FROM users 
WHERE apikey = $1
`

func (q *Queries) GetUserByApiKey(ctx context.Context, apikey string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByApiKey, apikey)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Apikey,
		&i.FamilyID,
		&i.Password,
	)
	return i, err
}

const getUserByName = `-- name: GetUserByName :one
select id, created_at, updated_at, name, apikey, family_id, password FROM users 
WHERE name=$1
`

func (q *Queries) GetUserByName(ctx context.Context, name string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByName, name)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Apikey,
		&i.FamilyID,
		&i.Password,
	)
	return i, err
}

const getUsersByFamily = `-- name: GetUsersByFamily :many
SELECT id, name FROM users
WHERE family_id=$1
ORDER BY created_at ASC
`

type GetUsersByFamilyRow struct {
	ID   string
	Name string
}

func (q *Queries) GetUsersByFamily(ctx context.Context, familyID sql.NullInt64) ([]GetUsersByFamilyRow, error) {
	rows, err := q.db.QueryContext(ctx, getUsersByFamily, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUsersByFamilyRow
	for rows.Next() {
		var i GetUsersByFamilyRow
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}