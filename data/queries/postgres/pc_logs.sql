-- name: ListPCLogsAfterCursor :many
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

-- name: ListPCLogsBeforeCursor :many
-- Возвращает записи в обратном порядке для эффективного использования индекса по id. Реверс делается в Go.
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

-- name: ListPCLogsFirstPage :many
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

-- name: CountPCLogs :one
SELECT COUNT(*)
FROM pc_logs pl
         JOIN pcs p ON p.id = pl.pc_id
WHERE p.user_id = @user_id
  AND pl.pc_id = @pc_id;

-- name: BatchInsertPCLogs :exec
INSERT INTO pc_logs (pc_id,
                     command_id,
                     received_at,
                     completed_at,
                     status,
                     error)
SELECT pc_id,
       command_id,
       received_at,
       completed_at,
       status,
       error
FROM UNNEST(@pc_ids::uuid[]), UNNEST(@command_ids::varchar[]), UNNEST(@received_ats::timestamptz[]), UNNEST(@completed_ats::timestamptz[]), UNNEST(@statuses::varchar[]), UNNEST(@errors::varchar[]);