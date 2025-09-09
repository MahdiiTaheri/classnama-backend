BEGIN;

CREATE TYPE attendance_status AS ENUM ('present', 'absent', 'late', 'excused');

-- attendance_records table
CREATE TABLE IF NOT EXISTS attendance_records (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    teacher_id BIGINT REFERENCES teachers(id) ON DELETE SET NULL,
    classroom_id BIGINT REFERENCES classrooms(id) ON DELETE SET NULL,
    date DATE NOT NULL,
    status attendance_status NOT NULL DEFAULT 'present',
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (student_id, date)
);

CREATE INDEX IF NOT EXISTS idx_attendance_student_date ON attendance_records(student_id, date);
CREATE INDEX IF NOT EXISTS idx_attendance_classroom_date ON attendance_records(classroom_id, date);

COMMIT;
