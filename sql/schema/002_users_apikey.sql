-- +goose Up
ALTER TABLE users 
ADD COLUMN apikey VARCHAR(64) 
DEFAULT encode(sha256(random()::text::bytea), 'hex')
UNIQUE 
NOT NULL;

-- +goose Down
ALTER TABLE users DROP COLUMN apikey;