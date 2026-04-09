-- name: GetPCsByUserID :many
SELECT id, user_id, slug, name, description, can_power_on
FROM pcs
WHERE user_id = $1;

-- name: GetPCBySlug :one
SELECT id, user_id, slug, name, description, can_power_on
FROM pcs
WHERE user_id = $1 AND slug = $2;

-- name: GetPCByID :one
SELECT id, user_id, slug, name, description, can_power_on
FROM pcs
WHERE id = $1 AND user_id = $2;

-- name: UpdatePC :one
UPDATE pcs
SET
    name        = COALESCE(sqlc.narg('name'), name),
    slug        = COALESCE(sqlc.narg('slug'), slug),
    description = COALESCE(sqlc.narg('description'), description),
    can_power_on = COALESCE(sqlc.narg('can_power_on'), can_power_on)
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id')
RETURNING *;