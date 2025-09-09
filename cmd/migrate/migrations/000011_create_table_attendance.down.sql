BEGIN;

DROP INDEX IF EXISTS idx_attendance_classroom_date;
DROP INDEX IF EXISTS idx_attendance_student_date;
DROP TABLE IF EXISTS attendance_records;
DROP TYPE IF EXISTS attendance_status;

COMMIT;
