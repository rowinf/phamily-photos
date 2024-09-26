-- name: CreateFamily :one
INSERT INTO families (id, created_at, updated_at, name, description)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetFamilyById :one
SELECT * FROM families
WHERE id = $1;

-- name: GetUserFamily :one
SELECT * FROM families AS f
WHERE id = (
    SELECT family_id
		FROM users as u
        WHERE u.id=$1
);