-- name: VarRefCreate :exec
INSERT INTO var_ref(
    env_id, name, comment, create_time, update_time, var_id
) VALUES (
    ?     , ?   , ?      , ?          , ?          , ?
);

-- name: VarRefDelete :execrows
DELETE FROM var_ref WHERE env_id = ? AND name = ?;

-- name: VarRefList :many
SELECT * FROM var_ref
WHERE env_id = ?
ORDER BY name ASC;

-- name: VarRefShow :one
SELECT *
FROM var_ref
WHERE env_id = ? AND name = ?;

-- name: VarRefListByVarID :many
SELECT env.name AS env_name, var_ref.* FROM var_ref
JOIN env ON var_ref.env_id = env.env_id
WHERE var_id = ?
ORDER BY var_ref.name ASC;

-- name: VarRefUpdate :execrows
UPDATE var_ref SET
    env_id = COALESCE(sqlc.narg('env_id'), env_id),
    name = COALESCE(sqlc.narg('name'), name),
    comment = COALESCE(sqlc.narg('comment'), comment),
    create_time = COALESCE(sqlc.narg('create_time'), create_time),
    update_time = COALESCE(sqlc.narg('update_time'), update_time),
    var_id = COALESCE(sqlc.narg('var_id'), var_id)
WHERE var_ref_id = sqlc.arg('var_ref_id');