package store

import (
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
}

func NewStorage(db *sql.DB) Storage {
	return Storage{}
}
