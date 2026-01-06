// Package pagination provides utilities for handling pagination in HTTP requests.
package pagination

import (
	"net/http"
	"strconv"
)

const (
	// DefaultPage is the default page number when not specified.
	DefaultPage = 1
	// DefaultPerPage is the default number of items per page.
	DefaultPerPage = 20
	// MaxPerPage is the maximum allowed items per page.
	MaxPerPage = 100
)

// Params contains pagination parameters extracted from an HTTP request.
type Params struct {
	Page    int
	PerPage int
	Offset  int
}

// FromRequest extracts pagination parameters from an HTTP request.
// It reads "page" and "per_page" query parameters and validates them.
func FromRequest(r *http.Request) Params {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	if page < 1 {
		page = DefaultPage
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}

	return Params{
		Page:    page,
		PerPage: perPage,
		Offset:  (page - 1) * perPage,
	}
}

// FromValues creates pagination params from explicit values.
// Useful for programmatic pagination without HTTP context.
func FromValues(page, perPage int) Params {
	if page < 1 {
		page = DefaultPage
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}

	return Params{
		Page:    page,
		PerPage: perPage,
		Offset:  (page - 1) * perPage,
	}
}

// Result contains pagination metadata for a response.
type Result struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// NewResult creates a pagination result from params and total count.
func NewResult(params Params, total int64) Result {
	totalPages := int((total + int64(params.PerPage) - 1) / int64(params.PerPage))
	if totalPages == 0 {
		totalPages = 1
	}

	return Result{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    params.Page < totalPages,
		HasPrev:    params.Page > 1,
	}
}
