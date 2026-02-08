-- Add completions column to var table
ALTER TABLE var ADD COLUMN completions TEXT NOT NULL DEFAULT '[]';

-- Drop and recreate vw_var_expanded to include enabled (from 007) and completions
DROP VIEW vw_var_expanded;
CREATE VIEW vw_var_expanded AS
SELECT
    var_id,
    env_id,
    (SELECT name FROM env WHERE env_id = var.env_id) AS env_name,
    name,
    value,
    comment,
    create_time,
    update_time,
    enabled,
    completions
FROM var;
