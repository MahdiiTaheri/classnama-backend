-- create index on email (unique)
CREATE UNIQUE INDEX idx_students_email ON students(email);

-- create index on class
CREATE INDEX idx_students_class ON students(class);

