-- Create enum for roles
CREATE TYPE exec_role AS ENUM ('admin', 'manager');

-- Create execs table
CREATE TABLE IF NOT EXISTS execs (
    id BIGSERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role exec_role NOT NULL DEFAULT 'manager',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
