-- name: EnvCreate :one
INSERT INTO env (
    name, comment, create_time, update_time, enabled
) VALUES (
    ?   , ?      , ?          , ?          , ?
)
RETURNING name, comment, create_time, update_time, enabled;

-- name: EnvDelete :execrows
DELETE FROM env WHERE name = ?;

-- name: EnvFindID :one
SELECT env_id FROM env WHERE name = ?;

-- name: EnvList :many
SELECT * FROM env
ORDER BY name ASC;

-- name: EnvShow :one
SELECT
    name, comment, create_time, update_time, enabled
FROM env
WHERE name = ?;

-- See https://docs.sqlc.dev/en/latest/howto/named_parameters.html#nullable-parameters
-- name: EnvUpdate :execrows
UPDATE env SET
    name = COALESCE(sqlc.narg('new_name'), name),
    comment = COALESCE(sqlc.narg('comment'), comment),
    create_time = COALESCE(sqlc.narg('create_time'), create_time),
    update_time = COALESCE(sqlc.narg('update_time'), update_time),
    enabled = COALESCE(sqlc.narg('enabled'), enabled)
WHERE name = sqlc.arg('name');
