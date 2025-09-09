package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MahdiiTaheri/classnama-backend/internal/store"
	"github.com/redis/go-redis/v9"
)

type StudentStore struct {
	rdb *redis.Client
}

const studentListTTL = time.Second * 30

// List cache
func (e *StudentStore) GetList(ctx context.Context, key string) ([]*store.Student, error) {
	data, err := e.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var students []*store.Student
	if err := json.Unmarshal(data, &students); err != nil {
		return nil, err
	}
	return students, nil
}

// SetList caches the student list
func (e *StudentStore) SetList(ctx context.Context, key string, students []*store.Student) error {
	data, err := json.Marshal(students)
	if err != nil {
		return err
	}
	return e.rdb.SetEx(ctx, key, data, studentListTTL).Err()
}

// GetByTeacher caches students for a specific teacher
func (s *StudentStore) GetByTeacher(ctx context.Context, teacherID int64) ([]*store.Student, error) {
	key := fmt.Sprintf("students:teacher:%d", teacherID)
	data, err := s.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var students []*store.Student
	if err := json.Unmarshal(data, &students); err != nil {
		return nil, err
	}
	return students, nil
}

func (s *StudentStore) SetByTeacher(ctx context.Context, teacherID int64, students []*store.Student) error {
	key := fmt.Sprintf("students:teacher:%d", teacherID)
	data, err := json.Marshal(students)
	if err != nil {
		return err
	}
	return s.rdb.SetEx(ctx, key, data, studentListTTL).Err()
}
