-- name: CreatePhoto :one
INSERT INTO photos (id, created_at, updated_at, modified_at, name, alt_text, url, thumb_url, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetPhotosByUser :many
SELECT * FROM photos as p 
JOIN users as u ON p.user_id = u.id 
WHERE u.id=$1
ORDER BY p.created_at DESC
LIMIT $2;
