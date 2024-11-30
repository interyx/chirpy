-- +goose Up
CREATE TABLE users(
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  email TEXT NOT NULL,
  unique(email)
);

-- +goose Down
DROP TABLE users;

