-- +goose Up
-- +goose StatementBegin
CREATE TABLE pcs
(
    id           UUID DEFAULT uuidv7() PRIMARY KEY,
    user_id      UUID          NOT NULL,
    slug         VARCHAR(300)  NOT NULL,
    name         VARCHAR(255)  NOT NULL,
    description  VARCHAR(1024) NOT NULL,
    can_power_on BOOLEAN       NOT NULL,

    UNIQUE (user_id, slug)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pcs;
-- +goose StatementEnd
