-- Drop the index first
DROP INDEX IF EXISTS idx_students_teacher_id;

-- Drop the column
ALTER TABLE students
DROP COLUMN IF EXISTS teacher_id;
