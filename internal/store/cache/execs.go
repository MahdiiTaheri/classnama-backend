package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MahdiiTaheri/classnama-backend/internal/store"
	"github.com/redis/go-redis/v9"
)

type ExecStore struct {
	rdb *redis.Client
}

const execListTTL = 30 * time.Second

// GetList returns cached exec list or nil
func (e *ExecStore) GetList(ctx context.Context, key string) ([]*store.Exec, error) {
	data, err := e.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var execs []*store.Exec
	if err := json.Unmarshal(data, &execs); err != nil {
		return nil, err
	}
	return execs, nil
}

// SetList caches the exec list
func (e *ExecStore) SetList(ctx context.Context, key string, execs []*store.Exec) error {
	data, err := json.Marshal(execs)
	if err != nil {
		return err
	}
	return e.rdb.SetEx(ctx, key, data, execListTTL).Err()
}
