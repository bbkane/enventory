-- rename tables
ALTER TABLE env_var RENAME TO var;
ALTER TABLE env_ref RENAME TO var_ref;

-- rename foreign keys
ALTER TABLE var_ref RENAME COLUMN env_var_id TO var_id;


-- rename primary keys
ALTER TABLE var RENAME COLUMN env_var_id TO var_id;
ALTER TABLE var_ref RENAME COLUMN env_ref_id TO var_ref_id;

-- rename other columns
-- (no other columns necessary this time)

-- drop and recreate views
DROP VIEW vw_env_env_ref_env_var_unique_name;
CREATE VIEW vw_env_var_var_ref_unique_name AS
SELECT e.env_id, vr.name
FROM env e JOIN var_ref vr ON e.env_id = vr.env_id
UNION ALL
SELECT e.env_id, v.name
FROM env e JOIN var v ON e.env_id = v.env_id;

DROP VIEW vw_env_var_referenced_name;
CREATE VIEW vw_var_expanded AS
SELECT
    var_id,
    env_id,
    (SELECT name FROM env WHERE env_id = var.env_id) AS env_name,
    name,
    value,
    comment,
    create_time,
    update_time
FROM var;

DROP VIEW vw_env_ref_referenced_name;
CREATE VIEW vw_var_ref_expanded AS
SELECT
    var_ref_id,
    env_id,
    (SELECT name FROM env WHERE env_id = var_ref.env_id) AS env_name,
    name,
    var_id,
    (SELECT name FROM var WHERE var_id = var_ref.var_id) AS ref_var_name,
    (SELECT env.name FROM env JOIN var ON env.env_id = var.env_id WHERE var.var_id = var_ref.var_id) AS ref_env_name,
    comment,
    create_time,
    update_time
FROM var_ref;

-- drop and recreate triggers
DROP TRIGGER tr_env_ref_insert_check_unique_name;
CREATE TRIGGER tr_var_ref_insert_check_unique_name
BEFORE INSERT ON var_ref
FOR EACH ROW
BEGIN
    SELECT RAISE(FAIL, 'name already exists in env')
    FROM
    vw_env_var_var_ref_unique_name
    WHERE env_id = NEW.env_id AND name = NEW.name;
END;

DROP TRIGGER tr_env_ref_update_check_unique_name;
CREATE TRIGGER tr_var_ref_update_check_unique_name
BEFORE UPDATE ON var_ref
FOR EACH ROW
BEGIN
    SELECT
        CASE
            WHEN OLD.env_id != NEW.env_id OR OLD.name != NEW.name THEN (
                SELECT RAISE(FAIL, 'name already exists in env')
                FROM vw_env_var_var_ref_unique_name
                WHERE env_id = NEW.env_id AND name = NEW.name
            )
            END;
        END;

DROP TRIGGER tr_env_var_insert_check_unique_name;
CREATE TRIGGER tr_var_insert_check_unique_name
BEFORE INSERT ON var
FOR EACH ROW
BEGIN
    SELECT RAISE(FAIL, 'name already exists in env')
    FROM
    vw_env_var_var_ref_unique_name
    WHERE env_id = NEW.env_id AND name = NEW.name;
END;

DROP TRIGGER tr_env_var_update_check_unique_name;
CREATE TRIGGER tr_var_update_check_unique_name
BEFORE UPDATE ON var
FOR EACH ROW
BEGIN
    SELECT
        CASE
            WHEN OLD.env_id != NEW.env_id OR OLD.name != NEW.name THEN (
                SELECT RAISE(FAIL, 'name already exists in env')
                FROM vw_env_var_var_ref_unique_name
                WHERE env_id = NEW.env_id AND name = NEW.name
            )
            END;
        END;

-- drop and recreate indexes
DROP INDEX ix_env_ref_env_id;
CREATE INDEX ix_var_ref_env_id ON var_ref(env_id);

DROP INDEX ix_env_ref_env_var_id;
CREATE INDEX ix_var_ref_var_id ON var_ref(var_id);

DROP INDEX ix_env_var_env_id;
CREATE INDEX ix_var_env_id ON var(env_id);