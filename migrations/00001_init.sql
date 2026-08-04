-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    login TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
