-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name, apikey, password, family_id)
VALUES ($1, NOW(), NOW(), $2, encode(sha256(random()::text::bytea), 'hex'), $3, $4)
RETURNING *;

-- name: GetUserByApiKey :one
SELECT * FROM users 
WHERE apikey = $1;

-- name: GetUserByName :one
select * from users 
where name=$1;
