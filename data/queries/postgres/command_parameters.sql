-- name: UserCommandParameters :many
SELECT cp.id, cp.command_id, cp.name, cp.description, cp.type
FROM command_parameters cp
         JOIN commands c ON c.id = cp.command_id
         JOIN pcs p ON p.id = c.pc_id
WHERE p.user_id = @user_id
  AND cp.command_id = @command_id;

-- name: CreateUserPcCommandParameters :copyfrom
INSERT INTO command_parameters(command_id, name, description, type)
VALUES ($1, $2, $3, $4);

-- name: SyncUserPcCommandParameters :many
WITH user_check AS (SELECT 1
                    FROM commands c
                             JOIN pcs p ON p.id = c.pc_id
                    WHERE c.id = @command_id::UUID
                      AND p.user_id = @user_id::UUID),
     upserted AS (
         INSERT INTO command_parameters (id, command_id, name, description, type)
             SELECT UNNEST(@ids::UUID[]),
                    @command_id::UUID,
                    UNNEST(@names::VARCHAR[]),
                    UNNEST(@descriptions::VARCHAR[]),
                    UNNEST(@types::SMALLINT[])
             WHERE EXISTS (SELECT 1 FROM user_check)
             ON CONFLICT (id) DO UPDATE
                 SET name = EXCLUDED.name,
                     description = EXCLUDED.description,
                     type = EXCLUDED.type
             RETURNING *),
     deleted AS (
         DELETE FROM command_parameters
             WHERE command_id = @command_id::UUID
                 AND id != ALL (@ids::UUID[])
                 AND EXISTS (SELECT 1 FROM user_check))
SELECT *
FROM upserted;