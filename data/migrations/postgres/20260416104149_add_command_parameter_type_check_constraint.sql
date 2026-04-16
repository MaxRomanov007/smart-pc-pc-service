-- +goose Up
-- +goose StatementBegin
UPDATE command_parameters SET type = 1 WHERE type < 1;
UPDATE command_parameters SET type = 3 WHERE type > 3;

ALTER TABLE command_parameters
    ADD CONSTRAINT ck_command_parameter_type CHECK ( type >= 1 AND type <= 3 );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE command_parameters
    DROP CONSTRAINT ck_command_parameter_type;
-- +goose StatementEnd
