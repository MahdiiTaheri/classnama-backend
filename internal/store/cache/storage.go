package cache

import (
	"context"

	"github.com/MahdiiTaheri/classnama-backend/internal/store"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	Students interface {
		GetList(context.Context, string) ([]*store.Student, error)
		SetList(context.Context, string, []*store.Student) error
		GetByTeacher(context.Context, int64) ([]*store.Student, error)
		SetByTeacher(context.Context, int64, []*store.Student) error
	}
	Teachers interface {
		GetList(context.Context, string) ([]*store.Teacher, error)
		SetList(context.Context, string, []*store.Teacher) error
	}
	Execs interface {
		GetList(context.Context, string) ([]*store.Exec, error)
		SetList(context.Context, string, []*store.Exec) error
	}
}

func NewRedisStorage(rdb *redis.Client) Storage {
	return Storage{
		Students: &StudentStore{rdb: rdb},
		Teachers: &TeacherStore{rdb: rdb},
		Execs:    &ExecStore{rdb: rdb},
	}
}
