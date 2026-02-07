CREATE TABLE env_var_ref (
    id INTEGER PRIMARY KEY,
    env_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    comment TEXT NOT NULL,
    create_time TEXT NOT NULL,
    update_time TEXT NOT NULL,
    env_var_local_id INTEGER NOT NULL,
    FOREIGN KEY (env_id) REFERENCES env(id) ON DELETE CASCADE,
    FOREIGN KEY (env_var_local_id) REFERENCES env_var_local(id) ON DELETE RESTRICT,
    UNIQUE(env_id, name)
);

CREATE INDEX env_var_ref_env_id_idx ON env_var_ref(env_id);

CREATE INDEX env_var_ref_env_var_local_id ON env_var_ref(env_var_local_id);

CREATE TRIGGER env_var_ref_name_insert AFTER INSERT ON env_var_ref
BEGIN
    INSERT INTO env_var_unique_name(env_id, name) VALUES (NEW.env_id, NEW.name);
END;

CREATE TRIGGER env_var_ref_name_update AFTER UPDATE ON env_var_ref
WHEN old.name != new.name
BEGIN
    UPDATE env_var_unique_name SET name = NEW.name WHERE name = OLD.name;
END;

CREATE TRIGGER env_var_ref_name_delete AFTER DELETE ON env_var_ref
BEGIN
    DELETE FROM env_var_unique_name WHERE name = OLD.name;
END;