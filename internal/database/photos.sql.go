// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: photos.sql

package database

import (
	"context"
	"database/sql"
	"time"
)

const createPhoto = `-- name: CreatePhoto :one
INSERT INTO photos (id, created_at, updated_at, modified_at, name, alt_text, url, thumb_url, user_id)
VALUES ($1, NOW(), NOW(), $2, $3, $4, $5, $6, $7)
RETURNING id, created_at, updated_at, modified_at, name, alt_text, url, thumb_url, user_id, TRUE AS is_my_photo
`

type CreatePhotoParams struct {
	ID         string
	ModifiedAt time.Time
	Name       string
	AltText    string
	Url        string
	ThumbUrl   string
	UserID     string
}

type CreatePhotoRow struct {
	ID         string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ModifiedAt time.Time
	Name       string
	AltText    string
	Url        string
	ThumbUrl   string
	UserID     string
	IsMyPhoto  bool
}

func (q *Queries) CreatePhoto(ctx context.Context, arg CreatePhotoParams) (CreatePhotoRow, error) {
	row := q.db.QueryRowContext(ctx, createPhoto,
		arg.ID,
		arg.ModifiedAt,
		arg.Name,
		arg.AltText,
		arg.Url,
		arg.ThumbUrl,
		arg.UserID,
	)
	var i CreatePhotoRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ModifiedAt,
		&i.Name,
		&i.AltText,
		&i.Url,
		&i.ThumbUrl,
		&i.UserID,
		&i.IsMyPhoto,
	)
	return i, err
}

const deletePhoto = `-- name: DeletePhoto :exec
DELETE FROM photos
    WHERE id=$1 AND user_id=$2
`

type DeletePhotoParams struct {
	ID     string
	UserID string
}

func (q *Queries) DeletePhoto(ctx context.Context, arg DeletePhotoParams) error {
	_, err := q.db.ExecContext(ctx, deletePhoto, arg.ID, arg.UserID)
	return err
}

const getPhoto = `-- name: GetPhoto :one
SELECT id, created_at, updated_at, modified_at, name, alt_text, url, thumb_url, user_id, p.user_id = $2 AS is_my_photo 
    FROM photos AS p
    WHERE id=$1 AND user_id=$2
`

type GetPhotoParams struct {
	ID     string
	UserID string
}

type GetPhotoRow struct {
	ID         string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ModifiedAt time.Time
	Name       string
	AltText    string
	Url        string
	ThumbUrl   string
	UserID     string
	IsMyPhoto  bool
}

func (q *Queries) GetPhoto(ctx context.Context, arg GetPhotoParams) (GetPhotoRow, error) {
	row := q.db.QueryRowContext(ctx, getPhoto, arg.ID, arg.UserID)
	var i GetPhotoRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ModifiedAt,
		&i.Name,
		&i.AltText,
		&i.Url,
		&i.ThumbUrl,
		&i.UserID,
		&i.IsMyPhoto,
	)
	return i, err
}

const getPhotosByUser = `-- name: GetPhotosByUser :many
SELECT p.id, p.created_at, p.updated_at, modified_at, p.name, alt_text, url, thumb_url, user_id, u.id, u.created_at, u.updated_at, u.name, apikey, family_id FROM photos AS p 
    JOIN users AS u ON p.user_id = u.id 
    WHERE u.id=$1
    ORDER BY p.modified_at DESC
    LIMIT $2
`

type GetPhotosByUserParams struct {
	ID    string
	Limit int32
}

type GetPhotosByUserRow struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ModifiedAt  time.Time
	Name        string
	AltText     string
	Url         string
	ThumbUrl    string
	UserID      string
	ID_2        string
	CreatedAt_2 time.Time
	UpdatedAt_2 time.Time
	Name_2      string
	Apikey      string
	FamilyID    sql.NullInt64
}

func (q *Queries) GetPhotosByUser(ctx context.Context, arg GetPhotosByUserParams) ([]GetPhotosByUserRow, error) {
	rows, err := q.db.QueryContext(ctx, getPhotosByUser, arg.ID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPhotosByUserRow
	for rows.Next() {
		var i GetPhotosByUserRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.ModifiedAt,
			&i.Name,
			&i.AltText,
			&i.Url,
			&i.ThumbUrl,
			&i.UserID,
			&i.ID_2,
			&i.CreatedAt_2,
			&i.UpdatedAt_2,
			&i.Name_2,
			&i.Apikey,
			&i.FamilyID,
		); err != nil {
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

const getPhotosByUserFamily = `-- name: GetPhotosByUserFamily :many
SELECT p.id, p.created_at, p.updated_at, p.modified_at, p.name, p.alt_text, p.url, p.thumb_url, p.user_id, u.name as user_name, p.user_id = $1 AS is_my_photo
FROM public.photos AS p
	JOIN users AS u ON p.user_id = u.id
	WHERE u.family_id = (
		SELECT family_id
            FROM users as u
            WHERE u.id=$1
	)
ORDER BY p.modified_at DESC
LIMIT $2
`

type GetPhotosByUserFamilyParams struct {
	UserID string
	Limit  int32
}

type GetPhotosByUserFamilyRow struct {
	ID         string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ModifiedAt time.Time
	Name       string
	AltText    string
	Url        string
	ThumbUrl   string
	UserID     string
	UserName   string
	IsMyPhoto  bool
}

func (q *Queries) GetPhotosByUserFamily(ctx context.Context, arg GetPhotosByUserFamilyParams) ([]GetPhotosByUserFamilyRow, error) {
	rows, err := q.db.QueryContext(ctx, getPhotosByUserFamily, arg.UserID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPhotosByUserFamilyRow
	for rows.Next() {
		var i GetPhotosByUserFamilyRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.ModifiedAt,
			&i.Name,
			&i.AltText,
			&i.Url,
			&i.ThumbUrl,
			&i.UserID,
			&i.UserName,
			&i.IsMyPhoto,
		); err != nil {
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
