-- 1. Add the teacher_id column with correct type (matches teachers.id)
ALTER TABLE students
ADD COLUMN teacher_id BIGINT NOT NULL;

-- 2. Add foreign key constraint
ALTER TABLE students
ADD CONSTRAINT students_teacher_id_fkey
FOREIGN KEY (teacher_id) REFERENCES teachers(id);

-- 3. Create index for performance
CREATE INDEX idx_students_teacher_id ON students(teacher_id);
