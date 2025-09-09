package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type AttendanceRecord struct {
	ID          int64     `json:"id"`
	StudentID   int64     `json:"student_id"`
	TeacherID   *int64    `json:"teacher_id,omitempty"`
	ClassroomID *int64    `json:"classroom_id,omitempty"`
	Date        time.Time `json:"date"`   // date part only
	Status      string    `json:"status"` // 'present','absent','late','excused'
	Note        *string   `json:"note,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type AttendanceStore struct {
	db *sql.DB
}

func NewAttendanceStore(db *sql.DB) *AttendanceStore {
	return &AttendanceStore{db: db}
}

// Mark inserts or updates a single attendance record (upsert by student_id+date).
func (s *AttendanceStore) Mark(ctx context.Context, rec *AttendanceRecord) error {
	if rec == nil {
		return fmt.Errorf("attendance record is nil")
	}
	// make sure date has no time component (set to midnight)
	rec.Date = rec.Date.UTC().Truncate(24 * time.Hour)

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		INSERT INTO attendance_records (student_id, teacher_id, classroom_id, date, status, note)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (student_id, date)
		DO UPDATE SET
		  teacher_id = EXCLUDED.teacher_id,
		  classroom_id = EXCLUDED.classroom_id,
		  status = EXCLUDED.status,
		  note = EXCLUDED.note
		RETURNING id, created_at
	`

	var teacherID interface{}
	if rec.TeacherID == nil {
		teacherID = nil
	} else {
		teacherID = *rec.TeacherID
	}
	var classroomID interface{}
	if rec.ClassroomID == nil {
		classroomID = nil
	} else {
		classroomID = *rec.ClassroomID
	}
	var note interface{}
	if rec.Note == nil || strings.TrimSpace(*rec.Note) == "" {
		note = nil
	} else {
		note = *rec.Note
	}

	err := s.db.QueryRowContext(ctx, query,
		rec.StudentID,
		teacherID,
		classroomID,
		rec.Date,
		rec.Status,
		note,
	).Scan(&rec.ID, &rec.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

// BulkMark marks attendance for many students in a single transaction.
// statuses is a map[studentID]status
func (s *AttendanceStore) BulkMark(ctx context.Context, classroomID int64, date time.Time, statuses map[int64]string) error {
	if len(statuses) == 0 {
		return nil
	}
	date = date.UTC().Truncate(24 * time.Hour)
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO attendance_records (student_id, teacher_id, classroom_id, date, status, note)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (student_id, date)
		DO UPDATE SET
		  classroom_id = EXCLUDED.classroom_id,
		  status = EXCLUDED.status,
		  note = EXCLUDED.note
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for sid, status := range statuses {
		// note left nil in bulk API - frontends can call Mark for notes
		if _, err := stmt.ExecContext(ctx, sid, nil, classroomID, date, status, nil); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// GetByStudent returns attendance records for a student between optional from/to (inclusive).
// Pass nil for from/to to get all.
func (s *AttendanceStore) GetByStudent(ctx context.Context, studentID int64, from, to *time.Time) ([]*AttendanceRecord, error) {
	args := []any{studentID}
	cond := "WHERE student_id = $1"
	i := 2
	if from != nil {
		args = append(args, from.UTC().Truncate(24*time.Hour))
		cond += fmt.Sprintf(" AND date >= $%d", i)
		i++
	}
	if to != nil {
		args = append(args, to.UTC().Truncate(24*time.Hour))
		cond += fmt.Sprintf(" AND date <= $%d", i)
		i++
	}
	query := fmt.Sprintf(`
		SELECT id, student_id, teacher_id, classroom_id, date, status, note, created_at
		FROM attendance_records
		%s
		ORDER BY date ASC
	`, cond)

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*AttendanceRecord{}
	for rows.Next() {
		var ar AttendanceRecord
		var teacher sql.NullInt64
		var classroom sql.NullInt64
		var note sql.NullString
		if err := rows.Scan(&ar.ID, &ar.StudentID, &teacher, &classroom, &ar.Date, &ar.Status, &note, &ar.CreatedAt); err != nil {
			return nil, err
		}
		if teacher.Valid {
			v := teacher.Int64
			ar.TeacherID = &v
		}
		if classroom.Valid {
			v := classroom.Int64
			ar.ClassroomID = &v
		}
		if note.Valid {
			n := note.String
			ar.Note = &n
		}
		out = append(out, &ar)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// GetByClassroomDate returns attendance for a classroom on a given date.
func (s *AttendanceStore) GetByClassroomDate(ctx context.Context, classroomID int64, date time.Time) ([]*AttendanceRecord, error) {
	date = date.UTC().Truncate(24 * time.Hour)
	query := `
		SELECT id, student_id, teacher_id, classroom_id, date, status, note, created_at
		FROM attendance_records
		WHERE classroom_id = $1 AND date = $2
		ORDER BY student_id ASC
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, classroomID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*AttendanceRecord{}
	for rows.Next() {
		var ar AttendanceRecord
		var teacher sql.NullInt64
		var classroom sql.NullInt64
		var note sql.NullString
		if err := rows.Scan(&ar.ID, &ar.StudentID, &teacher, &classroom, &ar.Date, &ar.Status, &note, &ar.CreatedAt); err != nil {
			return nil, err
		}
		if teacher.Valid {
			v := teacher.Int64
			ar.TeacherID = &v
		}
		if classroom.Valid {
			v := classroom.Int64
			ar.ClassroomID = &v
		}
		if note.Valid {
			n := note.String
			ar.Note = &n
		}
		out = append(out, &ar)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AttendanceStore) Delete(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	res, err := s.db.ExecContext(ctx, `DELETE FROM attendance_records WHERE id = $1`, id)
	if err != nil {
		return err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return ErrNotFound
	}
	return nil
}
