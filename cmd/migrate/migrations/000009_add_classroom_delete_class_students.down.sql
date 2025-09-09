-- Remove classroom_id column
ALTER TABLE students
DROP COLUMN classroom_id;

-- Re-add old class column
ALTER TABLE students
ADD COLUMN class VARCHAR(10) NOT NULL DEFAULT '1A';
