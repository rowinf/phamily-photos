-- name: CreatePhoto :one
INSERT INTO photos (id, created_at, updated_at, modified_at, name, alt_text, url, thumb_url, user_id, post_id)
VALUES ($1, NOW(), NOW(), $2, $3, $4, $5, $6, $7, $8)
RETURNING *, TRUE AS is_my_photo;

-- name: GetPhotosByUser :many
SELECT * FROM photos AS p 
    JOIN users AS u ON p.user_id = u.id 
    WHERE u.id=$1
    ORDER BY p.modified_at DESC
    LIMIT $2;

-- name: GetPostsByUserFamily :many
SELECT 
    p.id AS post_id,
    p.created_at AS post_created_at,
    p.updated_at AS post_updated_at,
    p.description AS post_description,
    u.name AS user_name,
    f.name AS family_name,
    ph.id AS photo_id,
    ph.name AS photo_name,
    ph.url AS photo_url,
    ph.thumb_url AS photo_thumb_url
FROM 
    posts p
JOIN 
    users u ON p.user_id = u.id
JOIN 
    families f ON p.family_id = f.id
LEFT JOIN 
    photos ph ON ph.post_id = p.id
WHERE
    p.family_id = $1
ORDER BY
    p.created_at DESC
LIMIT $2;

-- name: GetPostsByUserFamilyAggregated :many
SELECT
    p.id AS post_id,
    p.created_at AS post_created_at,
    p.updated_at AS post_updated_at,
    p.description AS post_description,
    u.name AS user_name,
    f.name AS family_name,
    json_arrayagg(json_build_object(
        'photo_id', ph.id,
        'photo_name', ph.name,
        'photo_url', ph.url,
        'photo_thumb_url', ph.thumb_url
    )) AS photos
FROM 
    posts p
JOIN 
    users u ON p.user_id = u.id
JOIN 
    families f ON u.family_id = f.id
LEFT JOIN 
    photos ph ON ph.post_id = p.id
WHERE 
    u.family_id = $1
GROUP BY 
    p.id, u.name, f.name
ORDER BY 
    p.created_at DESC
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
SELECT *, p.user_id = $2 AS is_my_photo, u.name AS user_name 
    FROM photos AS p
    JOIN users AS u ON p.user_id = u.id
    WHERE p.id=$1 AND user_id=$2;

-- name: DeletePhoto :exec
DELETE FROM photos
    WHERE id=$1 AND user_id=$2;

-- name: UpdatePhotosPostId :exec
UPDATE photos
SET post_id = $1
WHERE id = ANY($2::string[]);
