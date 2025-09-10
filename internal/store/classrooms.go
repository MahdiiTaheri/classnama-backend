package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Classroom struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Capacity  int64     `json:"capacity"`
	Grade     int64     `json:"grade"`
	TeacherID int64     `json:"teacher_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ClassroomStore interface {
	Create(ctx context.Context, classroom *Classroom) error
	GetByID(ctx context.Context, id int64) (*Classroom, error)
	GetAll(ctx context.Context, pq PaginatedQuery) ([]*Classroom, error)
	Update(ctx context.Context, classroom *Classroom) error
	Delete(ctx context.Context, id int64) error
}

type classroomStore struct {
	db *sql.DB
}

func NewClassroomStore(db *sql.DB) ClassroomStore {
	return &classroomStore{db: db}
}

func (s *classroomStore) Create(ctx context.Context, classroom *Classroom) error {
	query := `
		INSERT INTO classrooms (name, capacity, grade, teacher_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return s.db.QueryRowContext(ctx, query, classroom.Name, classroom.Capacity, classroom.Grade, classroom.TeacherID).
		Scan(&classroom.ID, &classroom.CreatedAt, &classroom.UpdatedAt)
}

func (s *classroomStore) GetByID(ctx context.Context, id int64) (*Classroom, error) {
	query := `
		SELECT id, name, capacity, grade, teacher_id, created_at, updated_at
		FROM classrooms
		WHERE id = $1
	`
	row := s.db.QueryRowContext(ctx, query, id)

	var c Classroom
	err := row.Scan(&c.ID, &c.Name, &c.Capacity, &c.Grade, &c.TeacherID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (s *classroomStore) GetAll(ctx context.Context, pq PaginatedQuery) ([]*Classroom, error) {
	columns := []string{"id", "name", "capacity", "grade", "created_at", "updated_at", "teacher_id"}
	searchCols := []string{"name"}

	query, args := BuildPaginatedQuery("classrooms", columns, pq, searchCols)

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	classrooms := []*Classroom{}
	for rows.Next() {
		var c Classroom
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Capacity,
			&c.Grade,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		classrooms = append(classrooms, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return classrooms, nil
}

func (s *classroomStore) Update(ctx context.Context, classroom *Classroom) error {
	query := `
		UPDATE classrooms
		SET name = $1, capacity = $2, grade = $3,teacher_id = $4 , updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`

	err := s.db.QueryRowContext(ctx, query,
		classroom.Name, classroom.Capacity, classroom.Grade, classroom.TeacherID, classroom.ID,
	).Scan(&classroom.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *classroomStore) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM classrooms WHERE id = $1`
	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
