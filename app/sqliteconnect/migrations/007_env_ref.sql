-- create env_ref table to reference envs from other envs

CREATE TABLE env_ref (
    env_ref_id INTEGER PRIMARY KEY,
    env_id INTEGER NOT NULL, -- env that contains us
    ref_env_id INTEGER NOT NULL, -- env that we reference
    comment TEXT NOT NULL,
    create_time TEXT NOT NULL,
    update_time TEXT NOT NULL,
    FOREIGN KEY(env_id) REFERENCES env(env_id) ON DELETE CASCADE,
    FOREIGN KEY(ref_env_id) REFERENCES env(env_id) ON DELETE RESTRICT,
    UNIQUE(env_id, ref_env_id)
) STRICT;

CREATE INDEX ix_env_ref_env_id ON env_ref(env_id);
CREATE INDEX ix_env_ref_ref_env_id ON env_ref(ref_env_id);

-- next let's create our env_export view, then add the triggers, then delete the old view and triggers
