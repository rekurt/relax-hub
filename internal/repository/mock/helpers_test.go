package mock

import (
	"testing"
)

func TestPaginate(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	tests := []struct {
		name       string
		items      []int
		page       int
		pageSize   int
		wantLen    int
		wantTotal  int64
		wantPages  int
		wantFirst  int
		wantPage   int
		wantPgSize int
	}{
		{
			name: "first page",
			items: items, page: 1, pageSize: 3,
			wantLen: 3, wantTotal: 10, wantPages: 4, wantFirst: 1, wantPage: 1, wantPgSize: 3,
		},
		{
			name: "middle page",
			items: items, page: 2, pageSize: 3,
			wantLen: 3, wantTotal: 10, wantPages: 4, wantFirst: 4, wantPage: 2, wantPgSize: 3,
		},
		{
			name: "last partial page",
			items: items, page: 4, pageSize: 3,
			wantLen: 1, wantTotal: 10, wantPages: 4, wantFirst: 10, wantPage: 4, wantPgSize: 3,
		},
		{
			name: "page beyond range",
			items: items, page: 5, pageSize: 3,
			wantLen: 0, wantTotal: 10, wantPages: 4, wantPage: 5, wantPgSize: 3,
		},
		{
			name: "empty items",
			items: []int{}, page: 1, pageSize: 10,
			wantLen: 0, wantTotal: 0, wantPages: 0, wantPage: 1, wantPgSize: 10,
		},
		{
			name: "nil items",
			items: nil, page: 1, pageSize: 10,
			wantLen: 0, wantTotal: 0, wantPages: 0, wantPage: 1, wantPgSize: 10,
		},
		{
			name: "defaults for invalid page and pageSize",
			items: items, page: 0, pageSize: 0,
			wantLen: 10, wantTotal: 10, wantPages: 1, wantFirst: 1, wantPage: 1, wantPgSize: 20,
		},
		{
			name: "negative page defaults to 1",
			items: items, page: -5, pageSize: 5,
			wantLen: 5, wantTotal: 10, wantPages: 2, wantFirst: 1, wantPage: 1, wantPgSize: 5,
		},
		{
			name: "exact fit",
			items: []int{1, 2, 3, 4}, page: 2, pageSize: 2,
			wantLen: 2, wantTotal: 4, wantPages: 2, wantFirst: 3, wantPage: 2, wantPgSize: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := paginate(tt.items, tt.page, tt.pageSize)

			if len(result.Items) != tt.wantLen {
				t.Errorf("Items len = %d, want %d", len(result.Items), tt.wantLen)
			}
			if result.TotalCount != tt.wantTotal {
				t.Errorf("TotalCount = %d, want %d", result.TotalCount, tt.wantTotal)
			}
			if result.TotalPages != tt.wantPages {
				t.Errorf("TotalPages = %d, want %d", result.TotalPages, tt.wantPages)
			}
			if result.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", result.Page, tt.wantPage)
			}
			if result.PageSize != tt.wantPgSize {
				t.Errorf("PageSize = %d, want %d", result.PageSize, tt.wantPgSize)
			}
			if tt.wantLen > 0 && result.Items[0] != tt.wantFirst {
				t.Errorf("first item = %d, want %d", result.Items[0], tt.wantFirst)
			}
		})
	}
}
