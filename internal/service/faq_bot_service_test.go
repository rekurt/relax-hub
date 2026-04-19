package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

type faqTestEnv struct {
	svc     service.FAQBotService
	faqRepo *mock.FAQRepo
}

func newFAQTestEnv() *faqTestEnv {
	faqRepo := mock.NewFAQRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewFAQBotService(faqRepo, log)
	return &faqTestEnv{
		svc:     svc,
		faqRepo: faqRepo,
	}
}

func validFAQ() *domain.FAQ {
	return &domain.FAQ{
		Category:  domain.FAQCategoryBooking,
		Question:  "Как забронировать баню?",
		Answer:    "Найдите подходящую баню через поиск и нажмите «Забронировать».",
		Keywords:  []string{"бронирование", "забронировать"},
		SortOrder: 1,
	}
}

// --- Create ---

func TestFAQBotService_Create_Success(t *testing.T) {
	env := newFAQTestEnv()
	faq := validFAQ()

	result, err := env.svc.CreateFAQ(context.Background(), faq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID == uuid.Nil {
		t.Error("expected FAQ ID to be set")
	}
	if !result.Active {
		t.Error("expected FAQ to be active")
	}
	if result.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestFAQBotService_Create_InvalidCategory(t *testing.T) {
	env := newFAQTestEnv()
	faq := validFAQ()
	faq.Category = "invalid"

	_, err := env.svc.CreateFAQ(context.Background(), faq)
	if err == nil {
		t.Fatal("expected error for invalid category")
	}
}

func TestFAQBotService_Create_EmptyQuestion(t *testing.T) {
	env := newFAQTestEnv()
	faq := validFAQ()
	faq.Question = ""

	_, err := env.svc.CreateFAQ(context.Background(), faq)
	if err == nil {
		t.Fatal("expected error for empty question")
	}
}

func TestFAQBotService_Create_EmptyAnswer(t *testing.T) {
	env := newFAQTestEnv()
	faq := validFAQ()
	faq.Answer = ""

	_, err := env.svc.CreateFAQ(context.Background(), faq)
	if err == nil {
		t.Fatal("expected error for empty answer")
	}
}

// --- Update ---

func TestFAQBotService_Update_Success(t *testing.T) {
	env := newFAQTestEnv()
	faq := validFAQ()
	created, _ := env.svc.CreateFAQ(context.Background(), faq)

	updated := &domain.FAQ{
		ID:        created.ID,
		Category:  domain.FAQCategoryPayment,
		Question:  "Обновлённый вопрос?",
		Answer:    "Обновлённый ответ.",
		Keywords:  []string{"оплата"},
		SortOrder: 2,
		Active:    true,
	}

	result, err := env.svc.UpdateFAQ(context.Background(), updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Category != domain.FAQCategoryPayment {
		t.Errorf("expected category payment, got %s", result.Category)
	}
	if result.Question != "Обновлённый вопрос?" {
		t.Errorf("expected updated question, got %s", result.Question)
	}
}

func TestFAQBotService_Update_NotFound(t *testing.T) {
	env := newFAQTestEnv()
	faq := &domain.FAQ{
		ID:       uuid.New(),
		Category: domain.FAQCategoryBooking,
		Question: "q",
		Answer:   "a",
	}

	_, err := env.svc.UpdateFAQ(context.Background(), faq)
	if err == nil {
		t.Fatal("expected error for not found FAQ")
	}
}

// --- Delete ---

func TestFAQBotService_Delete_Success(t *testing.T) {
	env := newFAQTestEnv()
	faq := validFAQ()
	created, _ := env.svc.CreateFAQ(context.Background(), faq)

	err := env.svc.DeleteFAQ(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = env.svc.GetFAQ(context.Background(), created.ID)
	if err == nil {
		t.Fatal("expected error after deletion")
	}
}

func TestFAQBotService_Delete_NotFound(t *testing.T) {
	env := newFAQTestEnv()
	err := env.svc.DeleteFAQ(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for not found FAQ")
	}
}

// --- Get ---

func TestFAQBotService_Get_Success(t *testing.T) {
	env := newFAQTestEnv()
	faq := validFAQ()
	created, _ := env.svc.CreateFAQ(context.Background(), faq)

	result, err := env.svc.GetFAQ(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Question != faq.Question {
		t.Errorf("expected question %s, got %s", faq.Question, result.Question)
	}
}

func TestFAQBotService_Get_NotFound(t *testing.T) {
	env := newFAQTestEnv()
	_, err := env.svc.GetFAQ(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for not found FAQ")
	}
}

// --- List ---

func TestFAQBotService_List_All(t *testing.T) {
	env := newFAQTestEnv()

	for i := 0; i < 3; i++ {
		faq := validFAQ()
		faq.Question = faq.Question + string(rune('0'+i))
		_, _ = env.svc.CreateFAQ(context.Background(), faq)
	}

	result, err := env.svc.ListFAQ(context.Background(), domain.FAQFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("expected 3 items, got %d", result.TotalCount)
	}
}

func TestFAQBotService_List_FilterByCategory(t *testing.T) {
	env := newFAQTestEnv()

	faq1 := validFAQ()
	faq1.Category = domain.FAQCategoryBooking
	_, _ = env.svc.CreateFAQ(context.Background(), faq1)

	faq2 := validFAQ()
	faq2.Category = domain.FAQCategoryPayment
	faq2.Question = "Как оплатить?"
	_, _ = env.svc.CreateFAQ(context.Background(), faq2)

	cat := domain.FAQCategoryPayment
	result, err := env.svc.ListFAQ(context.Background(), domain.FAQFilter{
		Category: &cat,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("expected 1 item, got %d", result.TotalCount)
	}
}

// --- Match ---

func TestFAQBotService_Match_Success(t *testing.T) {
	env := newFAQTestEnv()

	faq := validFAQ()
	_, _ = env.svc.CreateFAQ(context.Background(), faq)

	matches, err := env.svc.MatchFAQ(context.Background(), "бронирование")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("expected at least one match")
	}
	if matches[0].Score <= 0 {
		t.Error("expected positive match score")
	}
}

func TestFAQBotService_Match_NoResults(t *testing.T) {
	env := newFAQTestEnv()

	faq := validFAQ()
	_, _ = env.svc.CreateFAQ(context.Background(), faq)

	matches, err := env.svc.MatchFAQ(context.Background(), "xyzzy123nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestFAQBotService_Match_EmptyQuery(t *testing.T) {
	env := newFAQTestEnv()
	_, err := env.svc.MatchFAQ(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestFAQBotService_Match_InactiveFAQ(t *testing.T) {
	env := newFAQTestEnv()

	faq := validFAQ()
	created, _ := env.svc.CreateFAQ(context.Background(), faq)

	// Deactivate the FAQ
	created.Active = false
	_, _ = env.svc.UpdateFAQ(context.Background(), created)

	matches, err := env.svc.MatchFAQ(context.Background(), "бронирование")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for inactive FAQ, got %d", len(matches))
	}
}

// --- Seed ---

func TestFAQBotService_Seed_Success(t *testing.T) {
	env := newFAQTestEnv()

	err := env.svc.SeedDefaultFAQ(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, _ := env.svc.ListFAQ(context.Background(), domain.FAQFilter{Page: 1, PageSize: 50})
	if result.TotalCount == 0 {
		t.Fatal("expected seeded FAQ entries")
	}
	if result.TotalCount < 10 {
		t.Errorf("expected at least 10 seeded entries, got %d", result.TotalCount)
	}
}

func TestFAQBotService_Seed_Idempotent(t *testing.T) {
	env := newFAQTestEnv()

	_ = env.svc.SeedDefaultFAQ(context.Background())
	result1, _ := env.svc.ListFAQ(context.Background(), domain.FAQFilter{Page: 1, PageSize: 50})

	_ = env.svc.SeedDefaultFAQ(context.Background())
	result2, _ := env.svc.ListFAQ(context.Background(), domain.FAQFilter{Page: 1, PageSize: 50})

	if result1.TotalCount != result2.TotalCount {
		t.Errorf("seed is not idempotent: %d != %d", result1.TotalCount, result2.TotalCount)
	}
}

// --- Validation ---

func TestFAQ_Validate_QuestionTooLong(t *testing.T) {
	faq := validFAQ()
	faq.Question = string(make([]byte, 501))
	if err := faq.Validate(); err == nil {
		t.Error("expected error for question > 500 chars")
	}
}

func TestFAQ_Validate_AnswerTooLong(t *testing.T) {
	faq := validFAQ()
	faq.Answer = string(make([]byte, 5001))
	if err := faq.Validate(); err == nil {
		t.Error("expected error for answer > 5000 chars")
	}
}

func TestFAQ_Validate_TooManyKeywords(t *testing.T) {
	faq := validFAQ()
	faq.Keywords = make([]string, 21)
	for i := range faq.Keywords {
		faq.Keywords[i] = "kw"
	}
	if err := faq.Validate(); err == nil {
		t.Error("expected error for > 20 keywords")
	}
}
