package domain

import (
	"time"

	"github.com/google/uuid"
)

// FAQCategory categorizes FAQ entries for easier filtering.
type FAQCategory string

const (
	FAQCategoryBooking      FAQCategory = "booking"
	FAQCategoryPayment      FAQCategory = "payment"
	FAQCategoryCancellation FAQCategory = "cancellation"
	FAQCategoryWallet       FAQCategory = "wallet"
	FAQCategoryAccount      FAQCategory = "account"
	FAQCategoryGeneral      FAQCategory = "general"
)

func (c FAQCategory) IsValid() bool {
	switch c {
	case FAQCategoryBooking, FAQCategoryPayment, FAQCategoryCancellation,
		FAQCategoryWallet, FAQCategoryAccount, FAQCategoryGeneral:
		return true
	}
	return false
}

// FAQ represents a frequently asked question entry.
type FAQ struct {
	ID        uuid.UUID   `json:"id"`
	Category  FAQCategory `json:"category"`
	Question  string      `json:"question"`
	Answer    string      `json:"answer"`
	Keywords  []string    `json:"keywords"`
	SortOrder int         `json:"sort_order"`
	Active    bool        `json:"active"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

func (f *FAQ) Validate() error {
	if !f.Category.IsValid() {
		return ErrInvalidInput
	}
	if f.Question == "" || len(f.Question) > 500 {
		return ErrInvalidInput
	}
	if f.Answer == "" || len(f.Answer) > 5000 {
		return ErrInvalidInput
	}
	if len(f.Keywords) > 20 {
		return ErrInvalidInput
	}
	return nil
}

// FAQMatch represents a matched FAQ entry with a relevance score.
type FAQMatch struct {
	FAQ   FAQ     `json:"faq"`
	Score float64 `json:"score"`
}

// FAQFilter is used to filter FAQ entries in list queries.
type FAQFilter struct {
	Category *FAQCategory
	Active   *bool
	Page     int
	PageSize int
}
