-- name: GetLogsByPCID :many
SELECT
    cl.id,
    cl.command_id,
    c.name AS command_name,
    cl.received_at,
    cl.completed_at,
    cl.status,
    cl.error
FROM command_logs cl
         JOIN commands c ON c.id = cl.command_id
WHERE c.pc_id = $1
ORDER BY
    CASE WHEN sqlc.arg('order')::text = 'asc'  THEN cl.received_at END ASC,
    CASE WHEN sqlc.arg('order')::text = 'desc' THEN cl.received_at END DESC;

-- name: GetLogsByPCIDFirstPage :many
SELECT
    cl.id,
    cl.command_id,
    c.name AS command_name,
    cl.received_at,
    cl.completed_at,
    cl.status,
    cl.error
FROM command_logs cl
         JOIN commands c ON c.id = cl.command_id
WHERE c.pc_id = sqlc.arg('pc_id')
ORDER BY
    CASE WHEN sqlc.arg('order')::text = 'asc'  THEN cl.received_at END ASC,
    CASE WHEN sqlc.arg('order')::text = 'desc' THEN cl.received_at END DESC
LIMIT sqlc.arg('limit');

-- name: GetLogsByPCIDAfterCursor :many
SELECT
    cl.id,
    cl.command_id,
    c.name AS command_name,
    cl.received_at,
    cl.completed_at,
    cl.status,
    cl.error
FROM command_logs cl
         JOIN commands c ON c.id = cl.command_id
WHERE c.pc_id = sqlc.arg('pc_id')
  AND (
    CASE
        WHEN sqlc.arg('order')::text = 'asc'  THEN cl.received_at > (SELECT received_at FROM command_logs WHERE id = sqlc.arg('cursor')::uuid)
        WHEN sqlc.arg('order')::text = 'desc' THEN cl.received_at < (SELECT received_at FROM command_logs WHERE id = sqlc.arg('cursor')::uuid)
        END
    )
ORDER BY
    CASE WHEN sqlc.arg('order')::text = 'asc'  THEN cl.received_at END ASC,
    CASE WHEN sqlc.arg('order')::text = 'desc' THEN cl.received_at END DESC
LIMIT sqlc.arg('limit');

-- name: GetLogsByPCIDBeforeCursor :many
SELECT *
FROM (
         SELECT
             cl.id,
             cl.command_id,
             c.name AS command_name,
             cl.received_at,
             cl.completed_at,
             cl.status,
             cl.error
         FROM command_logs cl
                  JOIN commands c ON c.id = cl.command_id
         WHERE c.pc_id = sqlc.arg('pc_id')
           AND (
             CASE
                 WHEN sqlc.arg('order')::text = 'asc'  THEN cl.received_at < (SELECT received_at FROM command_logs WHERE id = sqlc.arg('cursor')::uuid)
                 WHEN sqlc.arg('order')::text = 'desc' THEN cl.received_at > (SELECT received_at FROM command_logs WHERE id = sqlc.arg('cursor')::uuid)
                 END
             )
         ORDER BY
             CASE WHEN sqlc.arg('order')::text = 'asc'  THEN cl.received_at END DESC,
             CASE WHEN sqlc.arg('order')::text = 'desc' THEN cl.received_at END ASC
         LIMIT sqlc.arg('limit')
     ) sub
ORDER BY
    CASE WHEN sqlc.arg('order')::text = 'asc'  THEN sub.received_at END ASC,
    CASE WHEN sqlc.arg('order')::text = 'desc' THEN sub.received_at END DESC;

-- name: CountLogsByPCID :one
SELECT COUNT(*)
FROM command_logs cl
         JOIN commands c ON c.id = cl.command_id
WHERE c.pc_id = $1;