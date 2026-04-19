-- name: UserPCs :many
SELECT id, user_id, slug, name, description, can_power_on
FROM pcs
WHERE user_id = @user_id;

-- name: UserPCBySlug :one
SELECT id, user_id, slug, name, description, can_power_on
FROM pcs
WHERE user_id = @user_id
  AND slug = @slug;

-- name: UserPCByID :one
SELECT id, user_id, slug, name, description, can_power_on
FROM pcs
WHERE user_id = @user_id
  AND id = @id;

-- name: UpdateUserPC :one
UPDATE pcs
SET name         = COALESCE(sqlc.narg('name'), name),
    slug         = COALESCE(sqlc.narg('slug'), slug),
    description  = COALESCE(sqlc.narg('description'), description),
    can_power_on = COALESCE(sqlc.narg('can_power_on'), can_power_on)
WHERE user_id = @user_id
  AND id = @id
RETURNING *;

-- name: CreateUserPC :one
INSERT INTO pcs(user_id, slug, name, description, can_power_on)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;