-- +goose Up
-- +goose StatementBegin
CREATE TABLE pc_logs
(
    id           UUID DEFAULT uuidv7() PRIMARY KEY,
    pc_id        UUID        NOT NULL REFERENCES pcs (id),
    command_id   VARCHAR(36) NOT NULL,
    received_at  TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ NOT NULL,
    status       VARCHAR(64) NOT NULL,
    error        VARCHAR(512)
);
CREATE INDEX idx_pc_logs_pc_id ON pc_logs(pc_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pc_logs;
-- +goose StatementEnd
