-- rename tables
ALTER TABLE env_var_local RENAME TO env_var;

ALTER TABLE env_var_ref RENAME TO env_ref;

-- rename foreign keys
ALTER TABLE env_ref RENAME COLUMN env_var_local_id TO env_var_id;

-- rename primary keys
ALTER TABLE env RENAME COLUMN id TO env_id;

ALTER TABLE env_ref RENAME COLUMN id TO env_ref_id;

ALTER TABLE env_var RENAME COLUMN id TO env_var_id;

ALTER TABLE keyring_entry RENAME COLUMN id to keyring_entry_id;

-- delete unique_name table and triggers
DROP TABLE env_var_unique_name;
DROP TRIGGER env_var_local_name_insert;
DROP TRIGGER env_var_local_name_update;
DROP TRIGGER env_var_local_name_delete;
DROP TRIGGER env_var_ref_name_insert;
DROP TRIGGER env_var_ref_name_update;
DROP TRIGGER env_var_ref_name_delete;

-- add new unique_name view and triggers
CREATE VIEW vw_env_env_ref_env_var_unique_name AS
SELECT e.env_id, er.name
FROM env e JOIN env_ref er ON e.env_id = er.env_id
UNION ALL
SELECT e.env_id, ev.name
FROM env e JOIN env_var ev ON e.env_id = ev.env_id
;

CREATE TRIGGER tr_env_ref_insert_check_unique_name
BEFORE INSERT ON env_ref
FOR EACH ROW
BEGIN
    SELECT RAISE(FAIL, 'name already exists in env')
    FROM vw_env_env_ref_env_var_unique_name
    WHERE env_id = NEW.env_id AND name = NEW.name;
END;

CREATE TRIGGER tr_env_ref_update_check_unique_name
BEFORE UPDATE ON env_ref
FOR EACH ROW
BEGIN
    SELECT RAISE(FAIL, 'name already exists in env')
    FROM vw_env_env_ref_env_var_unique_name
    WHERE env_id = NEW.env_id AND name = NEW.name;
END;

CREATE TRIGGER tr_env_var_insert_check_unique_name
BEFORE INSERT ON env_var
FOR EACH ROW
BEGIN
    SELECT RAISE(FAIL, 'name already exists in env')
    FROM vw_env_env_ref_env_var_unique_name
    WHERE env_id = NEW.env_id AND name = NEW.name;
END;

CREATE TRIGGER tr_env_var_update_check_unique_name
BEFORE UPDATE ON env_var
FOR EACH ROW
BEGIN
    SELECT RAISE(FAIL, 'name already exists in env')
    FROM vw_env_env_ref_env_var_unique_name
    WHERE env_id = NEW.env_id AND name = NEW.name;
END;

-- drop and recreate indexes with new names
DROP INDEX env_var_local_env_id_idx;
CREATE INDEX ix_env_var_env_id ON env_var(env_id);

DROP INDEX env_var_ref_env_id_idx;
CREATE INDEX ix_env_ref_env_id ON env_ref(env_id);

DROP INDEX env_var_ref_env_var_local_id;
CREATE INDEX ix_env_ref_env_var_id ON env_ref(env_var_id);


-- create views with names

-- DROP VIEW IF EXISTS vw_env_ref_referenced_name;
CREATE VIEW vw_env_ref_referenced_name AS
SELECT
    env_ref_id,
    env_id,
    (SELECT name FROM env WHERE env_id = env_ref.env_id) AS env_name,
    name,
    env_var_id,
    (SELECT name FROM env_var WHERE env_var_id = env_ref.env_var_id) AS ref_var_name,
    (SELECT env.name FROM env JOIN env_var ON env.env_id = env_var.env_id WHERE env_var.env_var_id = env_ref.env_var_id) AS ref_env_name,
    comment,
    create_time,
    update_time
FROM env_ref;
-- SELECT * FROM vw_env_ref_referenced_name;

CREATE VIEW vw_env_var_referenced_name AS
SELECT
    env_var_id,
    env_id,
	(SELECT name FROM env WHERE env_id = env_var.env_id) AS env_name,
    name,
    value,
    comment,
    create_time,
    update_time
FROM env_var;
-- SELECT * FROM vw_env_var_referenced_name;