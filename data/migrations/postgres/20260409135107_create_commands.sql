-- +goose Up
-- +goose StatementBegin
CREATE TABLE commands
(
    id          UUID DEFAULT uuidv7() PRIMARY KEY,
    pc_id       UUID          NOT NULL REFERENCES pcs (id) ON DELETE CASCADE,
    name        VARCHAR(255)  NOT NULL,
    description VARCHAR(1024) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS commands;
-- +goose StatementEnd
