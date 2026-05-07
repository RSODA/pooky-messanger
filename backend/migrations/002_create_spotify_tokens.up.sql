CREATE TABLE spotify_tokens (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    access_token TEXT,
    refresh_token TEXT,
    expires_at TIMESTAMP
);