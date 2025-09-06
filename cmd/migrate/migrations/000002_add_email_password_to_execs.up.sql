-- Add email column
ALTER TABLE execs
ADD COLUMN email VARCHAR(255) UNIQUE NOT NULL;

-- Add password_hash column
ALTER TABLE execs
ADD COLUMN password_hash TEXT NOT NULL;
