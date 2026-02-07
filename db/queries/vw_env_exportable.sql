-- name: EnvExportableList :many
SELECT name, enabled, value FROM vw_env_exportable
WHERE env_id = ?
ORDER BY type ASC, name ASC;