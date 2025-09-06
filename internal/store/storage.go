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
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Execs: &ExecStore{db},
	}
}
