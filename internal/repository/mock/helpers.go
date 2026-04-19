package mock

import "github.com/rekurt/relax-hub/internal/domain"

// paginate applies pagination to a pre-filtered slice and returns a PaginatedResult.
// It normalizes page/pageSize defaults and handles boundary conditions.
func paginate[T any](items []T, page, pageSize int) *domain.PaginatedResult[T] {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	total := int64(len(items))
	start := (page - 1) * pageSize
	if start >= len(items) {
		return &domain.PaginatedResult[T]{
			Items:      nil,
			TotalCount: total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}

	return &domain.PaginatedResult[T]{
		Items:      items[start:end],
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}
}
