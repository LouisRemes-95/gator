-- +goose Up
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    title TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    description TEXT,
    published_at TIMESTAMP NOT NULL,
    feed_id UUID NOT NULL,
    FOREIGN KEY (feed_id) REFERENCES feeds(id)
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;