-- name: GetCommandParametersByCommandID :many
SELECT id, command_id, name, description, type
FROM command_parameters
WHERE command_id = $1;