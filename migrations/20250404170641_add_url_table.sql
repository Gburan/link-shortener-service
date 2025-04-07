-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS urls (
    id           SERIAL PRIMARY KEY,
    original_url TEXT   NOT NULL UNIQUE,
    shorted_url  TEXT   NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS urls;
-- +goose StatementEnd
