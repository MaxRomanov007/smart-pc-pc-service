-- name: ListPCLogs :many
SELECT cl.id,
       cl.command_id,
       c.name AS command_name,
       cl.received_at,
       cl.completed_at,
       cl.status,
       cl.error
FROM command_logs cl
         JOIN commands c ON c.id = cl.command_id
         JOIN pcs p ON p.id = c.pc_id
WHERE p.user_id = @user_id
  AND c.pc_id = @pc_id
ORDER BY CASE WHEN sqlc.arg('order')::text = 'asc' THEN cl.id END ASC,
         CASE WHEN sqlc.arg('order')::text = 'desc' THEN cl.id END DESC;

-- name: ListPCLogsFirstPage :many
SELECT cl.id,
       cl.command_id,
       c.name AS command_name,
       cl.received_at,
       cl.completed_at,
       cl.status,
       cl.error
FROM command_logs cl
         JOIN commands c ON c.id = cl.command_id
         JOIN pcs p ON p.id = c.pc_id
WHERE p.user_id = @user_id
  AND c.pc_id = @pc_id
ORDER BY CASE WHEN sqlc.arg('order')::text = 'asc' THEN cl.id END ASC,
         CASE WHEN sqlc.arg('order')::text = 'desc' THEN cl.id END DESC
LIMIT sqlc.arg('limit');

-- name: ListPCLogsAfterCursor :many
SELECT cl.id,
       cl.command_id,
       c.name AS command_name,
       cl.received_at,
       cl.completed_at,
       cl.status,
       cl.error
FROM command_logs cl
         JOIN commands c ON c.id = cl.command_id
         JOIN pcs p ON p.id = c.pc_id
WHERE p.user_id = @user_id
  AND c.pc_id = @pc_id
  AND (
    CASE
        WHEN sqlc.arg('order')::text = 'asc' THEN cl.id > sqlc.arg('cursor')::uuid
        WHEN sqlc.arg('order')::text = 'desc' THEN cl.id < sqlc.arg('cursor')::uuid
        END
    )
ORDER BY CASE WHEN sqlc.arg('order')::text = 'asc' THEN cl.id END ASC,
         CASE WHEN sqlc.arg('order')::text = 'desc' THEN cl.id END DESC
LIMIT sqlc.arg('limit');

-- name: ListPCLogsBeforeCursor :many
SELECT cl.id,
       cl.command_id,
       c.name AS command_name,
       cl.received_at,
       cl.completed_at,
       cl.status,
       cl.error
FROM command_logs cl
         JOIN commands c ON c.id = cl.command_id
         JOIN pcs p ON p.id = c.pc_id
WHERE p.user_id = @user_id
  AND c.pc_id = @pc_id
  AND (
    CASE
        WHEN sqlc.arg('order')::text = 'asc' THEN cl.id < sqlc.arg('cursor')::uuid
        WHEN sqlc.arg('order')::text = 'desc' THEN cl.id > sqlc.arg('cursor')::uuid
        END
    )
ORDER BY CASE WHEN sqlc.arg('order')::text = 'asc' THEN cl.id END DESC,
         CASE WHEN sqlc.arg('order')::text = 'desc' THEN cl.id END ASC
LIMIT sqlc.arg('limit');

-- name: CountPCLogs :one
SELECT COUNT(*)
FROM command_logs cl
         JOIN commands c ON c.id = cl.command_id
         JOIN pcs p ON p.id = c.pc_id
WHERE p.user_id = @user_id
  AND c.pc_id = @pc_id;