package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Teacher struct {
	ID          int64     `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	Password    string    `json:"-"`
	Subject     string    `json:"subject"`
	PhoneNumber string    `json:"phone_number"`
	HireDate    time.Time `json:"hire_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TeacherStore struct {
	db *sql.DB
}

func (s *TeacherStore) Create(ctx context.Context, teacher *Teacher) error {
	query := `
		INSERT INTO teachers (first_name, last_name, email, password, subject, phone_number, hire_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx,
		query,
		teacher.FirstName,
		teacher.LastName,
		teacher.Email,
		teacher.Password,
		teacher.Subject,
		teacher.PhoneNumber,
		teacher.HireDate,
	).Scan(
		&teacher.ID,
		&teacher.CreatedAt,
		&teacher.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *TeacherStore) GetAll(ctx context.Context) ([]*Teacher, error) {
	query := `
		SELECT id, first_name, last_name, email, subject, phone_number, hire_date, created_at, updated_at
		FROM teachers
		ORDER BY id ASC
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teachers := []*Teacher{}
	for rows.Next() {
		var t Teacher
		if err := rows.Scan(
			&t.ID,
			&t.FirstName,
			&t.LastName,
			&t.Email,
			&t.Subject,
			&t.PhoneNumber,
			&t.HireDate,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		teachers = append(teachers, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return teachers, nil
}

func (s *TeacherStore) GetByID(ctx context.Context, id int64) (*Teacher, error) {
	query := `
		SELECT id, first_name, last_name, email, subject, phone_number, hire_date, created_at, updated_at
		FROM teachers
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var t Teacher
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.FirstName,
		&t.LastName,
		&t.Email,
		&t.Subject,
		&t.PhoneNumber,
		&t.HireDate,
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

func (s *StudentStore) GetByTeacherID(ctx context.Context, teacherID int64) ([]*Student, error) {
	query := `
		SELECT 
			id, first_name, last_name, email, password, phone_number, class, birth_date, address, parent_name, parent_phone_number, teacher_id, created_at, updated_at
		FROM students
		WHERE teacher_id = $1
		ORDER BY id ASC
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, teacherID)
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
			&s.Password,
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

func (s *TeacherStore) Update(ctx context.Context, teacher *Teacher) error {
	query := `
		UPDATE teachers
		SET first_name = $1,
		    last_name = $2,
		    email = $3,
		    subject = $4,
		    phone_number = $5,
		    hire_date = $6,
		    updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query,
		teacher.FirstName,
		teacher.LastName,
		teacher.Email,
		teacher.Subject,
		teacher.PhoneNumber,
		teacher.HireDate,
		teacher.ID,
	).Scan(&teacher.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *TeacherStore) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM teachers WHERE id = $1`

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
