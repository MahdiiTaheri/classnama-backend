-- Add email column
ALTER TABLE execs
ADD COLUMN email VARCHAR(255) UNIQUE NOT NULL;

-- Add password column
ALTER TABLE execs
ADD COLUMN password TEXT NOT NULL;
