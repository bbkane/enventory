CREATE VIEW vw_env_exportable AS
SELECT
    v.env_id,
    (SELECT name FROM env WHERE env_id = v.env_id) AS env_name,
    v.name,
    'var' AS type,
    v.comment,
    v.enabled,
    v.value,
    v.create_time,
    v.update_time
FROM var v

UNION ALL

SELECT
    vr.env_id,
    (SELECT name FROM env WHERE env_id = vr.env_id) AS env_name,
    vr.name,
    'var_ref' AS type,
    vr.comment,
    vr.enabled,
    (SELECT value FROM var WHERE var_id = vr.var_id) AS value,
    vr.create_time,
    vr.update_time
FROM var_ref vr;
