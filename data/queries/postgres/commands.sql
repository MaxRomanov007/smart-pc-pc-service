-- name: UserPCCommands :many
SELECT c.id, c.pc_id, c.name, c.description
FROM commands c
         JOIN pcs p ON p.id = c.pc_id
WHERE p.user_id = @user_id
  AND c.pc_id = @pc_id;

-- name: CreateUserPcCommand :one
INSERT INTO commands(pc_id, name, description)
SELECT pcs.id, @name, @description
FROM pcs
WHERE pcs.id = @pc_id
  AND pcs.user_id = @user_id
RETURNING *;

-- name: DeleteUserPcCommand :one
DELETE
FROM commands c
WHERE c.id = @id
  AND EXISTS(SELECT 1
             FROM pcs p
             WHERE p.id = c.pc_id
               AND p.id = @pc_id
               AND p.user_id = @user_id)
RETURNING *;