-- Teachers indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_teachers_email ON teachers(email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_teachers_phone ON teachers(phone_number);

-- Execs indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_execs_email ON execs(email);
CREATE INDEX IF NOT EXISTS idx_execs_role ON execs(role);
