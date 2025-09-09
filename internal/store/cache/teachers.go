package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MahdiiTaheri/classnama-backend/internal/store"
	"github.com/redis/go-redis/v9"
)

type TeacherStore struct {
	rdb *redis.Client
}

const teacherListTTL = time.Second * 30

// List cache
func (e *TeacherStore) GetList(ctx context.Context, key string) ([]*store.Teacher, error) {
	data, err := e.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var teachers []*store.Teacher
	if err := json.Unmarshal(data, &teachers); err != nil {
		return nil, err
	}
	return teachers, nil
}

// SetList caches the teacher list
func (e *TeacherStore) SetList(ctx context.Context, key string, teachers []*store.Teacher) error {
	data, err := json.Marshal(teachers)
	if err != nil {
		return err
	}
	return e.rdb.SetEx(ctx, key, data, teacherListTTL).Err()
}
