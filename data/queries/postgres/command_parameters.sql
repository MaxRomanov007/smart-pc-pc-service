-- name: UserCommandParameters :many
SELECT cp.id, cp.command_id, cp.name, cp.description, cp.type
FROM command_parameters cp
         JOIN commands c ON c.id = cp.command_id
         JOIN pcs p ON p.id = c.pc_id
WHERE p.user_id = @user_id
  AND cp.command_id = @command_id;