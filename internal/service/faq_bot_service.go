package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

const (
	faqMatchLimit = 3
	faqMinScore   = 0.15
)

// FAQBotService provides FAQ matching and admin CRUD for FAQ entries.
type FAQBotService interface {
	// Admin CRUD
	CreateFAQ(ctx context.Context, faq *domain.FAQ) (*domain.FAQ, error)
	UpdateFAQ(ctx context.Context, faq *domain.FAQ) (*domain.FAQ, error)
	DeleteFAQ(ctx context.Context, id uuid.UUID) error
	GetFAQ(ctx context.Context, id uuid.UUID) (*domain.FAQ, error)
	ListFAQ(ctx context.Context, filter domain.FAQFilter) (*domain.PaginatedResult[domain.FAQ], error)

	// Client-facing
	MatchFAQ(ctx context.Context, query string) ([]domain.FAQMatch, error)

	// Seed
	SeedDefaultFAQ(ctx context.Context) error
}

type faqBotService struct {
	faqRepo repository.FAQRepository
	logger  *logger.Logger
}

func NewFAQBotService(
	faqRepo repository.FAQRepository,
	log *logger.Logger,
) FAQBotService {
	return &faqBotService{
		faqRepo: faqRepo,
		logger:  log,
	}
}

func (s *faqBotService) CreateFAQ(ctx context.Context, faq *domain.FAQ) (*domain.FAQ, error) {
	if err := faq.Validate(); err != nil {
		return nil, err
	}

	faq.ID = uuid.New()
	now := time.Now()
	faq.CreatedAt = now
	faq.UpdatedAt = now
	faq.Active = true

	if err := s.faqRepo.Create(ctx, faq); err != nil {
		return nil, err
	}

	s.logger.Info("faq created", "faq_id", faq.ID, "category", faq.Category)
	return faq, nil
}

func (s *faqBotService) UpdateFAQ(ctx context.Context, faq *domain.FAQ) (*domain.FAQ, error) {
	existing, err := s.faqRepo.GetByID(ctx, faq.ID)
	if err != nil {
		return nil, err
	}

	existing.Category = faq.Category
	existing.Question = faq.Question
	existing.Answer = faq.Answer
	existing.Keywords = faq.Keywords
	existing.SortOrder = faq.SortOrder
	existing.Active = faq.Active
	existing.UpdatedAt = time.Now()

	if err := existing.Validate(); err != nil {
		return nil, err
	}

	if err := s.faqRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	s.logger.Info("faq updated", "faq_id", existing.ID)
	return existing, nil
}

func (s *faqBotService) DeleteFAQ(ctx context.Context, id uuid.UUID) error {
	if err := s.faqRepo.Delete(ctx, id); err != nil {
		return err
	}
	s.logger.Info("faq deleted", "faq_id", id)
	return nil
}

func (s *faqBotService) GetFAQ(ctx context.Context, id uuid.UUID) (*domain.FAQ, error) {
	return s.faqRepo.GetByID(ctx, id)
}

func (s *faqBotService) ListFAQ(ctx context.Context, filter domain.FAQFilter) (*domain.PaginatedResult[domain.FAQ], error) {
	return s.faqRepo.List(ctx, filter)
}

func (s *faqBotService) MatchFAQ(ctx context.Context, query string) ([]domain.FAQMatch, error) {
	if query == "" {
		return nil, domain.ErrInvalidInput
	}

	matches, err := s.faqRepo.SearchByKeywords(ctx, query, faqMatchLimit)
	if err != nil {
		return nil, err
	}

	// Filter out low-confidence matches
	var filtered []domain.FAQMatch
	for _, m := range matches {
		if m.Score >= faqMinScore {
			filtered = append(filtered, m)
		}
	}

	return filtered, nil
}

func (s *faqBotService) SeedDefaultFAQ(ctx context.Context) error {
	// Check if any FAQ entries exist
	result, err := s.faqRepo.List(ctx, domain.FAQFilter{Page: 1, PageSize: 1})
	if err != nil {
		return err
	}
	if result.TotalCount > 0 {
		return nil
	}

	seeds := []domain.FAQ{
		{
			Category:  domain.FAQCategoryBooking,
			Question:  "Как забронировать баню?",
			Answer:    "Откройте карточку объекта, выберите дату и непрерывный интервал свободных слотов, проверьте стоимость в checkout, укажите контакты и подтвердите бронь.",
			Keywords:  []string{"бронирование", "забронировать", "как забронировать", "слоты", "checkout", "записаться"},
			SortOrder: 1,
		},
		{
			Category:  domain.FAQCategoryBooking,
			Question:  "Можно ли изменить бронирование?",
			Answer:    "Да. В деталях бронирования можно изменить дату, время и количество гостей. Для активной брони доступно до трёх изменений, если визит ещё не начался.",
			Keywords:  []string{"изменить бронирование", "перенести", "поменять дату", "изменить время", "модификация"},
			SortOrder: 2,
		},
		{
			Category:  domain.FAQCategoryCancellation,
			Question:  "Как отменить бронирование?",
			Answer:    "Перейдите в «Мои бронирования», откройте нужную бронь и нажмите «Отменить». Перед подтверждением сервис покажет актуальные условия возврата по политике конкретного объекта.",
			Keywords:  []string{"отменить", "отмена", "отказаться", "возврат при отмене", "cancel", "отменить бронирование"},
			SortOrder: 1,
		},
		{
			Category:  domain.FAQCategoryCancellation,
			Question:  "Где посмотреть условия возврата?",
			Answer:    "Политика отмены показывается на странице объекта, в checkout и в деталях бронирования. Размер возврата зависит от того, сколько времени осталось до визита.",
			Keywords:  []string{"возврат", "деньги обратно", "вернуть деньги", "срок возврата", "refund", "условия возврата"},
			SortOrder: 2,
		},
		{
			Category:  domain.FAQCategoryPayment,
			Question:  "Какие способы оплаты поддерживаются?",
			Answer:    "В клиентском checkout доступны банковская карта, СБП, Apple Pay, Google Pay и оплата из кошелька. Если на кошельке не хватает суммы, можно оплатить комбинированно: часть с кошелька, часть картой.",
			Keywords:  []string{"оплата", "способ оплаты", "карта", "apple pay", "google pay", "сбп", "payment", "кошелёк"},
			SortOrder: 1,
		},
		{
			Category:  domain.FAQCategoryPayment,
			Question:  "Что делать, если оплата не проходит?",
			Answer:    "Проверьте лимиты карты и повторите попытку другим способом: СБП, другая карта или кошелёк. Если платёж всё равно не подтверждается, создайте обращение в поддержку и укажите номер брони.",
			Keywords:  []string{"ошибка оплаты", "не проходит оплата", "отклонён платёж", "проблема с оплатой", "платёж не прошёл"},
			SortOrder: 2,
		},
		{
			Category:  domain.FAQCategoryWallet,
			Question:  "Что такое кошелёк и зачем он нужен?",
			Answer:    "Кошелёк BANI хранит внутренний баланс для быстрой оплаты, возвратов и бонусных начислений. Его удобно использовать, когда вы бронируете регулярно и не хотите каждый раз вводить карту.",
			Keywords:  []string{"кошелёк", "баланс", "wallet", "зачем нужен кошелёк", "внутренний баланс"},
			SortOrder: 1,
		},
		{
			Category:  domain.FAQCategoryWallet,
			Question:  "Как пополнить кошелёк?",
			Answer:    "В разделе «Кошелёк» можно пополнить баланс картой, а затем использовать эти средства для полной или частичной оплаты бронирования.",
			Keywords:  []string{"пополнить кошелёк", "пополнить", "внести деньги", "wallet topup", "баланс"},
			SortOrder: 2,
		},
		{
			Category:  domain.FAQCategoryAccount,
			Question:  "Как изменить номер телефона?",
			Answer:    "Откройте профиль и обновите номер телефона через подтверждение по SMS. Это нужно, чтобы бронирования, уведомления и защита аккаунта оставались привязаны к актуальному номеру.",
			Keywords:  []string{"изменить телефон", "сменить номер", "новый телефон", "телефон", "sms"},
			SortOrder: 1,
		},
		{
			Category:  domain.FAQCategoryAccount,
			Question:  "Как удалить аккаунт?",
			Answer:    "В профиле доступен запрос на удаление аккаунта. После подтверждения начинается 30-дневный период восстановления, в течение которого удаление можно отменить.",
			Keywords:  []string{"удалить аккаунт", "удалить профиль", "удаление", "delete account", "деактивация"},
			SortOrder: 2,
		},
		{
			Category:  domain.FAQCategoryGeneral,
			Question:  "Как связаться с поддержкой?",
			Answer:    "Если быстрый FAQ не помог, откройте раздел «Поддержка» в кабинете клиента и создайте обращение. Для срочных вопросов используйте контакты, указанные в разделе помощи.",
			Keywords:  []string{"поддержка", "связаться", "помощь", "support", "контакт", "обращение"},
			SortOrder: 1,
		},
		{
			Category:  domain.FAQCategoryGeneral,
			Question:  "Как оставить отзыв?",
			Answer:    "После завершённого визита в деталях бронирования появляется действие «Оставить отзыв». Отзыв связан с реальной бронью, поэтому в профиле и карточке объекта показывается подтверждённый опыт.",
			Keywords:  []string{"отзыв", "оценка", "review", "оставить отзыв", "написать отзыв", "после визита"},
			SortOrder: 2,
		},
	}

	for i := range seeds {
		seeds[i].ID = uuid.New()
		now := time.Now()
		seeds[i].Active = true
		seeds[i].CreatedAt = now
		seeds[i].UpdatedAt = now
		if err := s.faqRepo.Create(ctx, &seeds[i]); err != nil {
			return err
		}
	}

	s.logger.Info("seeded default FAQ entries", "count", len(seeds))
	return nil
}
