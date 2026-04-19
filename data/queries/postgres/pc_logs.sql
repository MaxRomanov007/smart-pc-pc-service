-- name: UserPCLogsAfterCursor :many
SELECT pl.id,
       pl.command_id,
       c.name AS command_name,
       pl.received_at,
       pl.completed_at,
       pl.status,
       pl.error
FROM pc_logs pl
         JOIN pcs p ON p.id = pl.pc_id
         LEFT JOIN commands c ON c.id::text = pl.command_id
WHERE p.user_id = @user_id
  AND pl.pc_id = @pc_id
  AND (
    CASE
        WHEN sqlc.arg('order')::text = 'asc' THEN pl.id > sqlc.arg('cursor')::uuid
        WHEN sqlc.arg('order')::text = 'desc' THEN pl.id < sqlc.arg('cursor')::uuid
        END
    )
ORDER BY CASE WHEN sqlc.arg('order')::text = 'asc' THEN pl.id END ASC,
         CASE WHEN sqlc.arg('order')::text = 'desc' THEN pl.id END DESC
LIMIT sqlc.arg('limit');

-- name: UserPCLogsBeforeCursor :many
SELECT pl.id,
       pl.command_id,
       c.name AS command_name,
       pl.received_at,
       pl.completed_at,
       pl.status,
       pl.error
FROM pc_logs pl
         JOIN pcs p ON p.id = pl.pc_id
         LEFT JOIN commands c ON c.id::text = pl.command_id
WHERE p.user_id = @user_id
  AND pl.pc_id = @pc_id
  AND (
    CASE
        WHEN sqlc.arg('order')::text = 'asc' THEN pl.id < sqlc.arg('cursor')::uuid
        WHEN sqlc.arg('order')::text = 'desc' THEN pl.id > sqlc.arg('cursor')::uuid
        END
    )
ORDER BY CASE WHEN sqlc.arg('order')::text = 'asc' THEN pl.id END DESC,
         CASE WHEN sqlc.arg('order')::text = 'desc' THEN pl.id END ASC
LIMIT sqlc.arg('limit');

-- name: UserPCLogsFirstPage :many
SELECT pl.id,
       pl.command_id,
       c.name AS command_name,
       pl.received_at,
       pl.completed_at,
       pl.status,
       pl.error
FROM pc_logs pl
         JOIN pcs p ON p.id = pl.pc_id
         LEFT JOIN commands c ON c.id::text = pl.command_id
WHERE p.user_id = @user_id
  AND pl.pc_id = @pc_id
ORDER BY CASE WHEN sqlc.arg('order')::text = 'asc' THEN pl.id END ASC,
         CASE WHEN sqlc.arg('order')::text = 'desc' THEN pl.id END DESC
LIMIT sqlc.arg('limit');

-- name: UserPCLogsCount :one
SELECT COUNT(*)
FROM pc_logs pl
         JOIN pcs p ON p.id = pl.pc_id
WHERE p.user_id = @user_id
  AND pl.pc_id = @pc_id;

-- name: CreatePCLogs :copyfrom
INSERT INTO pc_logs (pc_id,
                     command_id,
                     received_at,
                     completed_at,
                     status,
                     error)
VALUES ($1, $2, $3, $4, $5, $6);