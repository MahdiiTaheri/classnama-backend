package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource conflict")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Execs interface {
		Create(context.Context, *Exec) error
		GetAll(context.Context) ([]*Exec, error)
		GetByID(context.Context, int64) (*Exec, error)
		Update(context.Context, *Exec) error
		Delete(context.Context, int64) error
	}
	Teachers interface {
		Create(context.Context, *Teacher) error
		GetAll(context.Context) ([]*Teacher, error)
		GetByID(context.Context, int64) (*Teacher, error)
		Update(context.Context, *Teacher) error
		Delete(context.Context, int64) error
	}
	Students interface {
		Create(context.Context, *Student) error
		GetAll(context.Context) ([]*Student, error)
		GetByID(context.Context, int64) (*Student, error)
		Update(context.Context, *Student) error
		Delete(context.Context, int64) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Execs:    &ExecStore{db},
		Teachers: &TeacherStore{db},
		Students: &StudentStore{db},
	}
}
