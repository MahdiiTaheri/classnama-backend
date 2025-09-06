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
	ID           uint      `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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

func (s *ExecStore) GetAll(ctx context.Context) ([]*Exec, error) {
	query := `
	SELECT id, first_name, last_name, email, role, created_at, updated_at
	FROM execs
	ORDER BY id ASC
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	execs := []*Exec{}
	for rows.Next() {
		var e Exec
		if err := rows.Scan(
			&e.ID,
			&e.FirstName,
			&e.LastName,
			&e.Email,
			&e.Role,
			&e.CreatedAt,
			&e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		execs = append(execs, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return execs, nil
}
