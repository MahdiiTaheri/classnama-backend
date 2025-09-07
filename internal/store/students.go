package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Student struct {
	ID                int64     `json:"id"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	Email             string    `json:"email"`
	Password          password  `json:"-"`
	PhoneNumber       *string   `json:"phone_number"`
	Class             string    `json:"class"`
	BirthDate         time.Time `json:"birth_date"`
	Address           string    `json:"address"`
	ParentName        string    `json:"parent_name"`
	ParentPhoneNumber string    `json:"parent_phone_number"`
	TeacherID         int64     `json:"teacher_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type StudentStore struct {
	db *sql.DB
}

func (s *StudentStore) Create(ctx context.Context, student *Student) error {
	query := `
		INSERT INTO students (first_name, last_name, email, password, phone_number, class, birth_date, address, parent_name, parent_phone_number, teacher_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx,
		query,
		student.FirstName,
		student.LastName,
		student.Email,
		student.Password.hash,
		student.PhoneNumber,
		student.Class,
		student.BirthDate,
		student.Address,
		student.ParentName,
		student.ParentPhoneNumber,
		student.TeacherID,
	).Scan(
		&student.ID,
		&student.CreatedAt,
		&student.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *StudentStore) GetAll(ctx context.Context, pq PaginatedQuery) ([]*Student, error) {
	query := `
		SELECT id, first_name, last_name, email, phone_number, class, birth_date,
		       address, parent_name, parent_phone_number, teacher_id, created_at, updated_at
		FROM students
	`

	// Sorting with whitelist
	if pq.SortBy != "" {
		switch pq.SortBy {
		case "id", "first_name", "last_name", "email", "class", "birth_date", "created_at", "updated_at":
			query += " ORDER BY " + pq.SortBy + " " + pq.Order
		default:
			query += " ORDER BY id ASC"
		}
	} else {
		query += " ORDER BY id ASC"
	}

	// Pagination
	query += " LIMIT $1 OFFSET $2"

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, pq.Limit, pq.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := []*Student{}
	for rows.Next() {
		var s Student
		if err := rows.Scan(
			&s.ID,
			&s.FirstName,
			&s.LastName,
			&s.Email,
			&s.PhoneNumber,
			&s.Class,
			&s.BirthDate,
			&s.Address,
			&s.ParentName,
			&s.ParentPhoneNumber,
			&s.TeacherID,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		students = append(students, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return students, nil
}

func (s *StudentStore) GetByID(ctx context.Context, id int64) (*Student, error) {
	query := `
		SELECT id, first_name, last_name, email, phone_number, class, birth_date, address, parent_name, parent_phone_number, teacher_id, created_at, updated_at
		FROM students
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var t Student
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.FirstName,
		&t.LastName,
		&t.Email,
		&t.PhoneNumber,
		&t.Class,
		&t.BirthDate,
		&t.Address,
		&t.ParentName,
		&t.ParentPhoneNumber,
		&t.TeacherID,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &t, nil
}

func (s *StudentStore) GetByEmail(ctx context.Context, email string) (*Student, error) {
	query := `
		SELECT id, first_name, last_name, email, phone_number, class, birth_date, address, parent_name, parent_phone_number, teacher_id, created_at, updated_at
		FROM students
		WHERE email = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var t Student
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&t.ID,
		&t.FirstName,
		&t.LastName,
		&t.Email,
		&t.PhoneNumber,
		&t.Class,
		&t.BirthDate,
		&t.Address,
		&t.ParentName,
		&t.ParentPhoneNumber,
		&t.TeacherID,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &t, nil
}

func (s *StudentStore) Update(ctx context.Context, student *Student) error {
	query := `
		UPDATE students
		SET first_name = $1,
		    last_name = $2,
		    email = $3,
		    phone_number = $4,
		    class = $5,
		    birth_date = $6,
		    address = $7,
		    parent_name = $8,
		    parent_phone_number = $9,
			teacher_id = $10,
		    updated_at = NOW()
		WHERE id = $11
		RETURNING updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query,
		student.FirstName,
		student.LastName,
		student.Email,
		student.PhoneNumber,
		student.Class,
		student.BirthDate,
		student.Address,
		student.ParentName,
		student.ParentPhoneNumber,
		student.TeacherID,
		student.ID,
	).Scan(&student.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *StudentStore) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM students WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
