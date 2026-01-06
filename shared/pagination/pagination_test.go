package pagination

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromRequest(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		wantPage    int
		wantPerPage int
		wantOffset  int
	}{
		{
			name:        "default values when no params",
			queryParams: "",
			wantPage:    1,
			wantPerPage: 20,
			wantOffset:  0,
		},
		{
			name:        "custom page and per_page",
			queryParams: "?page=3&per_page=50",
			wantPage:    3,
			wantPerPage: 50,
			wantOffset:  100,
		},
		{
			name:        "page only",
			queryParams: "?page=5",
			wantPage:    5,
			wantPerPage: 20,
			wantOffset:  80,
		},
		{
			name:        "per_page only",
			queryParams: "?per_page=30",
			wantPage:    1,
			wantPerPage: 30,
			wantOffset:  0,
		},
		{
			name:        "invalid page defaults to 1",
			queryParams: "?page=-1&per_page=10",
			wantPage:    1,
			wantPerPage: 10,
			wantOffset:  0,
		},
		{
			name:        "invalid per_page defaults to 20",
			queryParams: "?page=2&per_page=0",
			wantPage:    2,
			wantPerPage: 20,
			wantOffset:  20,
		},
		{
			name:        "per_page exceeds max, capped to 100",
			queryParams: "?page=1&per_page=200",
			wantPage:    1,
			wantPerPage: 100,
			wantOffset:  0,
		},
		{
			name:        "non-numeric values default",
			queryParams: "?page=abc&per_page=xyz",
			wantPage:    1,
			wantPerPage: 20,
			wantOffset:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test"+tt.queryParams, nil)
			params := FromRequest(req)

			if params.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", params.Page, tt.wantPage)
			}
			if params.PerPage != tt.wantPerPage {
				t.Errorf("PerPage = %d, want %d", params.PerPage, tt.wantPerPage)
			}
			if params.Offset != tt.wantOffset {
				t.Errorf("Offset = %d, want %d", params.Offset, tt.wantOffset)
			}
		})
	}
}

func TestFromValues(t *testing.T) {
	tests := []struct {
		name        string
		page        int
		perPage     int
		wantPage    int
		wantPerPage int
		wantOffset  int
	}{
		{
			name:        "valid values",
			page:        2,
			perPage:     25,
			wantPage:    2,
			wantPerPage: 25,
			wantOffset:  25,
		},
		{
			name:        "page less than 1 defaults to 1",
			page:        0,
			perPage:     20,
			wantPage:    1,
			wantPerPage: 20,
			wantOffset:  0,
		},
		{
			name:        "per_page less than 1 defaults to 20",
			page:        1,
			perPage:     -5,
			wantPage:    1,
			wantPerPage: 20,
			wantOffset:  0,
		},
		{
			name:        "per_page exceeds max, capped to 100",
			page:        1,
			perPage:     150,
			wantPage:    1,
			wantPerPage: 100,
			wantOffset:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := FromValues(tt.page, tt.perPage)

			if params.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", params.Page, tt.wantPage)
			}
			if params.PerPage != tt.wantPerPage {
				t.Errorf("PerPage = %d, want %d", params.PerPage, tt.wantPerPage)
			}
			if params.Offset != tt.wantOffset {
				t.Errorf("Offset = %d, want %d", params.Offset, tt.wantOffset)
			}
		})
	}
}

func TestNewResult(t *testing.T) {
	tests := []struct {
		name           string
		params         Params
		total          int64
		wantTotalPages int
		wantHasNext    bool
		wantHasPrev    bool
	}{
		{
			name:           "first page with more pages",
			params:         Params{Page: 1, PerPage: 10, Offset: 0},
			total:          25,
			wantTotalPages: 3,
			wantHasNext:    true,
			wantHasPrev:    false,
		},
		{
			name:           "middle page",
			params:         Params{Page: 2, PerPage: 10, Offset: 10},
			total:          25,
			wantTotalPages: 3,
			wantHasNext:    true,
			wantHasPrev:    true,
		},
		{
			name:           "last page",
			params:         Params{Page: 3, PerPage: 10, Offset: 20},
			total:          25,
			wantTotalPages: 3,
			wantHasNext:    false,
			wantHasPrev:    true,
		},
		{
			name:           "single page",
			params:         Params{Page: 1, PerPage: 10, Offset: 0},
			total:          5,
			wantTotalPages: 1,
			wantHasNext:    false,
			wantHasPrev:    false,
		},
		{
			name:           "empty result",
			params:         Params{Page: 1, PerPage: 10, Offset: 0},
			total:          0,
			wantTotalPages: 1,
			wantHasNext:    false,
			wantHasPrev:    false,
		},
		{
			name:           "exact page boundary",
			params:         Params{Page: 1, PerPage: 10, Offset: 0},
			total:          20,
			wantTotalPages: 2,
			wantHasNext:    true,
			wantHasPrev:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewResult(tt.params, tt.total)

			if result.TotalPages != tt.wantTotalPages {
				t.Errorf("TotalPages = %d, want %d", result.TotalPages, tt.wantTotalPages)
			}
			if result.HasNext != tt.wantHasNext {
				t.Errorf("HasNext = %v, want %v", result.HasNext, tt.wantHasNext)
			}
			if result.HasPrev != tt.wantHasPrev {
				t.Errorf("HasPrev = %v, want %v", result.HasPrev, tt.wantHasPrev)
			}
			if result.Total != tt.total {
				t.Errorf("Total = %d, want %d", result.Total, tt.total)
			}
			if result.Page != tt.params.Page {
				t.Errorf("Page = %d, want %d", result.Page, tt.params.Page)
			}
			if result.PerPage != tt.params.PerPage {
				t.Errorf("PerPage = %d, want %d", result.PerPage, tt.params.PerPage)
			}
		})
	}
}
