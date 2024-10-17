-- name: GetPost :one
SELECT * FROM posts
WHERE id=$1
LIMIT 1;

-- name: CreatePost :one
INSERT INTO posts (description, featured_photo_id, user_id, family_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetPostsByUserFamily :many
-- SELECT p.*, u.name as user_name, p.user_id = $1 AS is_my_photo
