-- +goose Up
-- +goose StatementBegin
CREATE TABLE command_parameters
(
    id          UUID DEFAULT uuidv7() PRIMARY KEY,
    command_id  UUID          NOT NULL REFERENCES commands (id),
    name        VARCHAR(255)  NOT NULL,
    description VARCHAR(1024) NOT NULL,
    type        SMALLINT      NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS command_parameters;
-- +goose StatementEnd
