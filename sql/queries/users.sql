-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name, apikey, password, family_id)
VALUES ($1, NOW(), NOW(), $2, encode(sha256(random()::text::bytea), 'hex'), $3, $4)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users 
WHERE ID = $1;

-- name: GetUserByName :one
select * FROM users 
WHERE name=$1;

-- name: GetUsersByFamily :many
SELECT id, name FROM users
WHERE family_id=$1
ORDER BY created_at ASC;
