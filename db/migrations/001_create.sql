CREATE TABLE env (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    comment TEXT NOT NULL,
    create_time TEXT NOT NULL,
    update_time TEXT NOT NULL,
    UNIQUE(name)
) STRICT;

-- Create a table to keep all the env_names unique to the owning env

CREATE TABLE env_var_unique_name (
    id INTEGER PRIMARY KEY,
    env_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    FOREIGN KEY (env_id) REFERENCES env(id) ON DELETE CASCADE,
    UNIQUE(env_id, name)
) STRICT;

CREATE INDEX env_var_unique_name_env_id_idx ON env_var_unique_name(env_id);

-- Create env_var_local and associated triggers

CREATE TABLE env_var_local (
    id INTEGER PRIMARY KEY,
    env_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    comment TEXT NOT NULL,
    create_time TEXT NOT NULL,
    update_time TEXT NOT NULL,
    value TEXT NOT NULL,
    FOREIGN KEY (env_id) REFERENCES env(id) ON DELETE CASCADE,
    UNIQUE(env_id, name)
) STRICT;

CREATE INDEX env_var_local_env_id_idx ON env_var_local(env_id);

CREATE TRIGGER env_var_local_name_insert AFTER INSERT ON env_var_local
BEGIN
    INSERT INTO env_var_unique_name(env_id, name) VALUES (NEW.env_id, NEW.name);
END;

CREATE TRIGGER env_var_local_name_update AFTER UPDATE ON env_var_local
WHEN old.name != new.name
BEGIN
    UPDATE env_var_unique_name SET name = NEW.name WHERE name = OLD.name;
END;

CREATE TRIGGER env_var_local_name_delete AFTER DELETE ON env_var_local
BEGIN
    DELETE FROM env_var_unique_name WHERE name = OLD.name;
END;

-- Create keyring_entry

CREATE TABLE keyring_entry (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    comment TEXT NOT NULL,
    create_time TEXT NOT NULL,
    update_time TEXT NOT NULL,
    UNIQUE(name)
) STRICT;

