package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource conflict")
	QueryTimeoutDuration = time.Second * 5
)

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

func (p *password) Check(text string) bool {
	if p == nil || p.hash == nil {
		return false
	}
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(text))
	return err == nil
}

type Storage struct {
	Execs interface {
		Create(context.Context, *Exec) error
		GetAll(context.Context, PaginatedQuery) ([]*Exec, error)
		GetByID(context.Context, int64) (*Exec, error)
		GetByEmail(context.Context, string) (*Exec, error)
		Update(context.Context, *Exec) error
		Delete(context.Context, int64) error
	}
	Teachers interface {
		Create(context.Context, *Teacher) error
		GetAll(context.Context, PaginatedQuery) ([]*Teacher, error)
		GetByID(context.Context, int64) (*Teacher, error)
		GetByEmail(context.Context, string) (*Teacher, error)
		Update(context.Context, *Teacher) error
		Delete(context.Context, int64) error
	}
	Students interface {
		Create(context.Context, *Student) error
		GetAll(context.Context, PaginatedQuery) ([]*Student, error)
		GetByID(context.Context, int64) (*Student, error)
		GetByEmail(context.Context, string) (*Student, error)
		Update(context.Context, *Student) error
		Delete(context.Context, int64) error
		GetByTeacherID(ctx context.Context, teacherID int64) ([]*Student, error)
	}
	Classrooms interface {
		Create(context.Context, *Classroom) error
		GetAll(context.Context, PaginatedQuery) ([]*Classroom, error)
		GetByID(context.Context, int64) (*Classroom, error)
		Update(context.Context, *Classroom) error
		Delete(context.Context, int64) error
	}
	Attendance interface {
		Mark(context.Context, *AttendanceRecord) error
		BulkMark(context.Context, int64, time.Time, map[int64]string) error
		GetByStudent(context.Context, int64, *time.Time, *time.Time) ([]*AttendanceRecord, error)
		GetByClassroomDate(context.Context, int64, time.Time) ([]*AttendanceRecord, error)
		Delete(context.Context, int64) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Execs:      &ExecStore{db},
		Teachers:   &TeacherStore{db},
		Students:   &StudentStore{db},
		Classrooms: &classroomStore{db},
		Attendance: &AttendanceStore{db},
	}
}
