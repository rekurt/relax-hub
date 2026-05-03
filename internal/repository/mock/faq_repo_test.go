package mock

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

func TestFAQRepo_FullCoverage(t *testing.T) {
	ctx := context.Background()
	repo := NewFAQRepo()

	t.Run("CRUD", func(t *testing.T) {
		faq := &domain.FAQ{
			Category: domain.FAQCategoryBooking,
			Question: "Как отменить бронь?",
			Answer:   "В личном кабинете в разделе Мои брони.",
			Keywords: []string{"отмена", "бронь", "cancel"},
			Active:   true,
		}
		if err := repo.Create(ctx, faq); err != nil {
			t.Fatal(err)
		}

		got, err := repo.GetByID(ctx, faq.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Question != faq.Question {
			t.Errorf("Question: %q", got.Question)
		}
		if _, err := repo.GetByID(ctx, uuid.New()); err != domain.ErrFAQNotFound {
			t.Errorf("GetByID missing: %v", err)
		}

		got.Answer = "Нажмите Отменить."
		if err := repo.Update(ctx, got); err != nil {
			t.Fatal(err)
		}
		if err := repo.Update(ctx, &domain.FAQ{ID: uuid.New()}); err != domain.ErrFAQNotFound {
			t.Errorf("Update missing: %v", err)
		}

		if err := repo.Delete(ctx, faq.ID); err != nil {
			t.Fatal(err)
		}
		if err := repo.Delete(ctx, faq.ID); err != domain.ErrFAQNotFound {
			t.Errorf("Delete twice: %v", err)
		}
	})

	t.Run("ListWithFilters", func(t *testing.T) {
		fresh := NewFAQRepo()
		_ = fresh.Create(ctx, &domain.FAQ{Category: domain.FAQCategoryBooking, Question: "q1", Answer: "a", Active: true})
		_ = fresh.Create(ctx, &domain.FAQ{Category: domain.FAQCategoryPayment, Question: "q2", Answer: "a", Active: true})
		_ = fresh.Create(ctx, &domain.FAQ{Category: domain.FAQCategoryPayment, Question: "q3", Answer: "a", Active: false})

		page, err := fresh.List(ctx, domain.FAQFilter{Page: 1, PageSize: 10})
		if err != nil {
			t.Fatal(err)
		}
		if page.TotalCount != 3 {
			t.Errorf("List total: %d", page.TotalCount)
		}

		cat := domain.FAQCategoryPayment
		page, _ = fresh.List(ctx, domain.FAQFilter{Category: &cat, Page: 1, PageSize: 10})
		if page.TotalCount != 2 {
			t.Errorf("Category filter: %d, want 2", page.TotalCount)
		}

		active := true
		page, _ = fresh.List(ctx, domain.FAQFilter{Active: &active, Page: 1, PageSize: 10})
		if page.TotalCount != 2 {
			t.Errorf("Active filter: %d, want 2", page.TotalCount)
		}

		// pagination defaults & out-of-range
		page, _ = fresh.List(ctx, domain.FAQFilter{Page: 0, PageSize: 0})
		if page.Page != 1 || page.PageSize != 20 {
			t.Errorf("List defaults: %+v", page)
		}
		oor, _ := fresh.List(ctx, domain.FAQFilter{Page: 99, PageSize: 1})
		if len(oor.Items) != 0 {
			t.Errorf("Out-of-range list: %d", len(oor.Items))
		}
	})

	t.Run("SearchByKeywords", func(t *testing.T) {
		fresh := NewFAQRepo()
		active := &domain.FAQ{
			Category: domain.FAQCategoryBooking,
			Question: "Как отменить бронь?",
			Answer:   "В личном кабинете.",
			Keywords: []string{"cancel", "refund"},
			Active:   true,
		}
		_ = fresh.Create(ctx, active)

		_ = fresh.Create(ctx, &domain.FAQ{
			Category: domain.FAQCategoryPayment,
			Question: "Как пополнить кошелёк?",
			Answer:   "Через оплату.",
			Keywords: []string{"wallet", "topup"},
			Active:   true,
		})

		_ = fresh.Create(ctx, &domain.FAQ{
			Category: domain.FAQCategoryGeneral,
			Question: "Inactive q",
			Answer:   "Inactive a",
			Keywords: []string{"cancel"},
			Active:   false,
		})

		// Match by question text
		matches, err := fresh.SearchByKeywords(ctx, "отменить", 5)
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) != 1 {
			t.Errorf("Question match: %d, want 1", len(matches))
		}
		if matches[0].Score != 0.8 {
			t.Errorf("Question score: %v", matches[0].Score)
		}

		// Match by keyword
		matches, _ = fresh.SearchByKeywords(ctx, "cancel", 5)
		if len(matches) != 1 {
			t.Errorf("Keyword match: %d, want 1 (inactive excluded)", len(matches))
		}

		// No match
		matches, _ = fresh.SearchByKeywords(ctx, "xyzzy_no_match", 5)
		if len(matches) != 0 {
			t.Errorf("No match: %d", len(matches))
		}

		// Default limit
		_ = fresh.Create(ctx, &domain.FAQ{Category: domain.FAQCategoryBooking, Question: "another cancel q", Answer: "a", Active: true})
		_ = fresh.Create(ctx, &domain.FAQ{Category: domain.FAQCategoryBooking, Question: "third cancel q", Answer: "a", Active: true})
		_ = fresh.Create(ctx, &domain.FAQ{Category: domain.FAQCategoryBooking, Question: "fourth cancel q", Answer: "a", Active: true})
		matches, _ = fresh.SearchByKeywords(ctx, "cancel", 0)
		if len(matches) > 3 {
			t.Errorf("Default limit: %d", len(matches))
		}
	})

	t.Run("ListActiveByCategory", func(t *testing.T) {
		fresh := NewFAQRepo()
		_ = fresh.Create(ctx, &domain.FAQ{Category: domain.FAQCategoryAccount, Question: "q1", Answer: "a", Active: true})
		_ = fresh.Create(ctx, &domain.FAQ{Category: domain.FAQCategoryAccount, Question: "q2", Answer: "a", Active: false})
		_ = fresh.Create(ctx, &domain.FAQ{Category: domain.FAQCategoryWallet, Question: "q3", Answer: "a", Active: true})

		got, err := fresh.ListActiveByCategory(ctx, domain.FAQCategoryAccount)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Errorf("ListActiveByCategory: %d, want 1", len(got))
		}

		empty, _ := fresh.ListActiveByCategory(ctx, domain.FAQCategoryCancellation)
		if len(empty) != 0 {
			t.Errorf("Empty category: %d", len(empty))
		}
	})
}
