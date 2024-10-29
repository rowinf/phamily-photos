-- +goose Up
CREATE TABLE photos(
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    modified_at TIMESTAMP NOT NULL,
    name TEXT NOT NULL,
    alt_text TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    thumb_url TEXT UNIQUE NOT NULL,
    user_id TEXT REFERENCES users (id) ON DELETE CASCADE NOT NULL
);

-- +goose Down
DROP TABLE photos;