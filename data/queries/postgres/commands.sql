-- name: GetCommandsByPCID :many
SELECT id, pc_id, name, description
FROM commands
WHERE pc_id = $1;