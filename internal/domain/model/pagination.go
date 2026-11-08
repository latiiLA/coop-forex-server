package model

import "time"

// PageQuery is 1-based. Limit 0 means return all matched rows (no skip/limit).
type PageQuery struct {
	Page     int
	Limit    int
	DateFrom *time.Time
	DateTo   *time.Time
}

type PaginatedRequests struct {
	Items []Request `json:"items"`
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
}

type RequestStatusSummary struct {
	Total      int64            `json:"total"`
	ByStatus   map[string]int64 `json:"by_status"`
}

func NormalizePageQuery(page, limit int) PageQuery {
	if page < 1 {
		page = 1
	}
	if limit < 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	return PageQuery{Page: page, Limit: limit}
}
