package store

import (
	"net/http"
	"strconv"
)

// PaginatedQuery holds pagination and sorting params.
type PaginatedQuery struct {
	Limit  int    `json:"limit" validate:"gte=1,lte=100"`
	Offset int    `json:"offset" validate:"gte=0"`
	SortBy string `json:"sort_by" validate:"omitempty"`
	Order  string `json:"order" validate:"oneof=asc desc"`
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

	return pq, nil
}
