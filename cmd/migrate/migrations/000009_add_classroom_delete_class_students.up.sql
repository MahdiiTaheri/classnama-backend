ALTER TABLE students
DROP COLUMN class;

ALTER TABLE students
ADD COLUMN classroom_id BIGINT REFERENCES classrooms(id) ON DELETE SET NULL;
