-- name: CreatePhoto :one
INSERT INTO photos (id, created_at, updated_at, modified_at, name, alt_text, url, thumb_url, user_id)
VALUES ($1, NOW(), NOW(), $2, $3, $4, $5, $6, $7)
RETURNING *, TRUE AS is_my_photo;

-- name: GetPhotosByUser :many
SELECT * FROM photos AS p 
    JOIN users AS u ON p.user_id = u.id 
    WHERE u.id=$1
    ORDER BY p.modified_at DESC
    LIMIT $2;

-- name: GetPhotosByUserFamily :many
SELECT p.*, u.name as user_name, p.user_id = $1 AS is_my_photo
FROM public.photos AS p
	JOIN users AS u ON p.user_id = u.id
	WHERE u.family_id = (
		SELECT family_id
            FROM users as u
            WHERE u.id=$1
	)
ORDER BY p.modified_at DESC
LIMIT $2;

-- name: GetPhoto :one
SELECT *, p.user_id = $2 AS is_my_photo 
    FROM photos AS p
    WHERE id=$1 AND user_id=$2;

-- name: DeletePhoto :exec
DELETE FROM photos
    WHERE id=$1 AND user_id=$2;