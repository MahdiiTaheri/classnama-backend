package store

import (
	"context"
	"database/sql"
	"time"
)

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
)

type Exec struct {
	ID        uint      `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ExecStore struct {
	db *sql.DB
}

func (s *ExecStore) Create(ctx context.Context, exec *Exec) error {
	query := `
	INSERT INTO execs (first_name, last_name, role)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx,
		query,
		exec.FirstName,
		exec.LastName,
		exec.Role,
	).Scan(
		&exec.ID,
		&exec.CreatedAt,
		&exec.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}
