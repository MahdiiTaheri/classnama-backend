package cache

import (
	"context"
	"fmt"
	"sort"
)

// ListGetter fetches the list from DB
type ListGetter[T any] func(ctx context.Context) ([]*T, error)

// GetListWithCache tries Redis first, fetches from DB on miss, sets cache
func GetListWithCache[T any](
	ctx context.Context,
	rdb interface {
		GetList(context.Context, string) ([]*T, error)
		SetList(context.Context, string, []*T) error
	},
	prefix string,
	params map[string]any,
	fetcher ListGetter[T],
) ([]*T, error) {
	key := buildCacheKey(prefix, params)

	// Try cache
	if cached, err := rdb.GetList(ctx, key); err == nil && cached != nil {
		return cached, nil
	}

	// Fetch from DB
	list, err := fetcher(ctx)
	if err != nil {
		return nil, err
	}

	// Set cache
	_ = rdb.SetList(ctx, key, list)
	return list, nil
}

// buildCacheKey returns a deterministic key from params
func buildCacheKey(prefix string, params map[string]any) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	key := prefix + ":"
	for i, k := range keys {
		if i > 0 {
			key += "&"
		}
		key += fmt.Sprintf("%s=%v", k, params[k])
	}
	return key
}
