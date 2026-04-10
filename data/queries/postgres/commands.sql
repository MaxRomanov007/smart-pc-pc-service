-- name: ListPCCommands :many
SELECT c.id, c.pc_id, c.name, c.description
FROM commands c
         JOIN pcs p ON p.id = c.pc_id
WHERE p.user_id = @user_id
  AND c.pc_id = @pc_id;