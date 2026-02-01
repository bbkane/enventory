-- Add enabled column to env table
ALTER TABLE env ADD COLUMN enabled INTEGER NOT NULL DEFAULT 1;

-- Add enabled column to var table
ALTER TABLE var ADD COLUMN enabled INTEGER NOT NULL DEFAULT 1;

-- Add enabled column to var_ref table
ALTER TABLE var_ref ADD COLUMN enabled INTEGER NOT NULL DEFAULT 1;
