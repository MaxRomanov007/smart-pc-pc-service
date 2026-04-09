-- +goose Up
-- +goose StatementBegin
CREATE TABLE command_logs
(
    id           UUID DEFAULT uuidv7() PRIMARY KEY,
    command_id   UUID        NOT NULL REFERENCES commands (id),
    received_at  TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ NOT NULL,
    status       VARCHAR(64) NOT NULL,
    error        VARCHAR(512)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS command_logs;
-- +goose StatementEnd
