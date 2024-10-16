// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: photos.sql

package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

const createPhoto = `-- name: CreatePhoto :one
INSERT INTO photos (id, created_at, updated_at, modified_at, name, alt_text, url, thumb_url, user_id, post_id)
VALUES ($1, NOW(), NOW(), $2, $3, $4, $5, $6, $7, $8)
RETURNING id, created_at, updated_at, modified_at, name, alt_text, url, thumb_url, user_id, post_id, TRUE AS is_my_photo
`

type CreatePhotoParams struct {
	ID         string
	ModifiedAt time.Time
	Name       string
	AltText    string
	Url        string
	ThumbUrl   string
	UserID     string
	PostID     sql.NullInt64
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
	PostID     sql.NullInt64
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
		arg.PostID,
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
		&i.PostID,
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
SELECT p.id, p.created_at, p.updated_at, modified_at, p.name, alt_text, url, thumb_url, user_id, post_id, u.id, u.created_at, u.updated_at, u.name, apikey, family_id, password, p.user_id = $2 AS is_my_photo, u.name AS user_name 
    FROM photos AS p
    JOIN users AS u ON p.user_id = u.id
    WHERE p.id=$1 AND user_id=$2
`

type GetPhotoParams struct {
	ID     string
	UserID string
}

type GetPhotoRow struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ModifiedAt  time.Time
	Name        string
	AltText     string
	Url         string
	ThumbUrl    string
	UserID      string
	PostID      sql.NullInt64
	ID_2        string
	CreatedAt_2 time.Time
	UpdatedAt_2 time.Time
	Name_2      string
	Apikey      string
	FamilyID    sql.NullInt64
	Password    string
	IsMyPhoto   bool
	UserName    string
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
		&i.PostID,
		&i.ID_2,
		&i.CreatedAt_2,
		&i.UpdatedAt_2,
		&i.Name_2,
		&i.Apikey,
		&i.FamilyID,
		&i.Password,
		&i.IsMyPhoto,
		&i.UserName,
	)
	return i, err
}

const getPhotosByUser = `-- name: GetPhotosByUser :many
SELECT p.id, p.created_at, p.updated_at, modified_at, p.name, alt_text, url, thumb_url, user_id, post_id, u.id, u.created_at, u.updated_at, u.name, apikey, family_id, password FROM photos AS p 
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
	PostID      sql.NullInt64
	ID_2        string
	CreatedAt_2 time.Time
	UpdatedAt_2 time.Time
	Name_2      string
	Apikey      string
	FamilyID    sql.NullInt64
	Password    string
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
			&i.PostID,
			&i.ID_2,
			&i.CreatedAt_2,
			&i.UpdatedAt_2,
			&i.Name_2,
			&i.Apikey,
			&i.FamilyID,
			&i.Password,
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
SELECT p.id, p.created_at, p.updated_at, p.modified_at, p.name, p.alt_text, p.url, p.thumb_url, p.user_id, p.post_id, u.name as user_name, p.user_id = $1 AS is_my_photo
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
	PostID     sql.NullInt64
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
			&i.PostID,
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

const updatePhotosPostId = `-- name: UpdatePhotosPostId :exec
UPDATE photos
SET post_id = $1
WHERE id = ANY($2::string[])
`

type UpdatePhotosPostIdParams struct {
	PostID  sql.NullInt64
	Column2 []string
}

func (q *Queries) UpdatePhotosPostId(ctx context.Context, arg UpdatePhotosPostIdParams) error {
	_, err := q.db.ExecContext(ctx, updatePhotosPostId, arg.PostID, pq.Array(arg.Column2))
	return err
}
