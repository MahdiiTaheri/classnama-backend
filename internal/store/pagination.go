package store

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// PaginatedQuery holds pagination and sorting params.
type PaginatedQuery struct {
	Limit  int    `json:"limit" validate:"gte=1,lte=50,omitempty"`
	Offset int    `json:"offset" validate:"gte=0,omitempty"`
	SortBy string `json:"sort_by" validate:"omitempty"`
	Order  string `json:"order" validate:"oneof=asc desc,omitempty"`
	Search string `json:"search" validate:"max=72,omitempty"`
}

// Parse extracts pagination + sorting from query params.
func (pq PaginatedQuery) Parse(r *http.Request) (PaginatedQuery, error) {
	qs := r.URL.Query()

	limit := qs.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return pq, nil
		}

		pq.Limit = l
	}

	offset := qs.Get("offset")
	if offset != "" {
		l, err := strconv.Atoi(offset)
		if err != nil {
			return pq, nil
		}

		pq.Offset = l
	}

	sortBy := qs.Get("sort_by")
	if sortBy != "" {
		pq.SortBy = sortBy
	}

	if ord := qs.Get("order"); ord != "" {
		if ord == "asc" || ord == "desc" {
			pq.Order = ord
		}
	}

	if search := qs.Get("search"); search != "" {
		pq.Search = search
	}

	return pq, nil
}

func BuildPaginatedQuery(
	table string,
	columns []string,
	pq PaginatedQuery,
	searchColumns []string,
) (string, []any) {
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), table)
	args := []any{}
	argPos := 1 // keeps track of $1, $2, ...

	// Search
	if pq.Search != "" && len(searchColumns) > 0 {
		where := []string{}
		for _, col := range searchColumns {
			where = append(where, fmt.Sprintf("%s ILIKE $%d", col, argPos))
		}
		query += " WHERE " + strings.Join(where, " OR ")
		args = append(args, "%"+pq.Search+"%")
		argPos++
	}

	// Sorting
	if pq.SortBy != "" {
		query += " ORDER BY " + pq.SortBy
		if pq.Order == "desc" {
			query += " DESC"
		} else {
			query += " ASC"
		}
	} else {
		query += " ORDER BY id ASC"
	}

	// Pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, pq.Limit, pq.Offset)

	return query, args
}
