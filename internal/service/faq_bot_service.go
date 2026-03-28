package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

const (
	faqMatchLimit  = 3
	faqMinScore    = 0.15
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
			Answer:    "Найдите подходящую баню через поиск, выберите удобную дату и время, укажите количество гостей и нажмите «Забронировать». Оплата происходит онлайн при подтверждении.",
			Keywords:  []string{"бронирование", "забронировать", "букинг", "как забронировать", "записаться"},
			SortOrder: 1,
		},
		{
			Category:  domain.FAQCategoryBooking,
			Question:  "Можно ли изменить бронирование?",
			Answer:    "Да, вы можете изменить дату, время, длительность и количество гостей до 3 раз. Перейдите в раздел «Мои бронирования», выберите нужное и нажмите «Изменить». Разница в цене будет пересчитана автоматически.",
			Keywords:  []string{"изменить бронирование", "перенести", "поменять дату", "изменить время"},
			SortOrder: 2,
		},
		{
			Category:  domain.FAQCategoryCancellation,
			Question:  "Как отменить бронирование?",
			Answer:    "Перейдите в «Мои бронирования», выберите нужное и нажмите «Отменить». Размер возврата зависит от политики отмены объекта: гибкая (100% при отмене за 24ч), умеренная (100% за 72ч) или строгая (100% за 7 дней). Подробности указаны на странице бронирования.",
			Keywords:  []string{"отменить", "отмена", "отказаться", "возврат при отмене", "cancel"},
			SortOrder: 1,
		},
		{
			Category:  domain.FAQCategoryCancellation,
			Question:  "Когда я получу возврат после отмены?",
			Answer:    "При отмене средства возвращаются на кошелёк мгновенно с бонусом 5%, либо на банковскую карту в течение 3-10 рабочих дней. Выберите удобный вариант при отмене.",
			Keywords:  []string{"возврат", "деньги обратно", "вернуть деньги", "срок возврата", "refund"},
			SortOrder: 2,
		},
		{
			Category:  domain.FAQCategoryPayment,
			Question:  "Какие способы оплаты поддерживаются?",
			Answer:    "Мы принимаем банковские карты (Visa, Mastercard, МИР), Apple Pay, Google Pay, СБП и оплату из кошелька приложения. Также можно комбинировать кошелёк и карту.",
			Keywords:  []string{"оплата", "способ оплаты", "карта", "apple pay", "google pay", "сбп", "payment"},
			SortOrder: 1,
		},
		{
			Category:  domain.FAQCategoryPayment,
			Question:  "Оплата не проходит, что делать?",
			Answer:    "Убедитесь, что на карте достаточно средств и она поддерживает онлайн-платежи. Попробуйте другую карту или оплатите через СБП. Если проблема сохраняется, обратитесь в поддержку.",
			Keywords:  []string{"ошибка оплаты", "не проходит оплата", "отклонён платёж", "проблема с оплатой"},
			SortOrder: 2,
		},
		{
			Category:  domain.FAQCategoryWallet,
			Question:  "Что такое кошелёк и как его пополнить?",
			Answer:    "Кошелёк — внутренний баланс для быстрой оплаты. Пополните его картой (от 500 до 30 000 ₽ за раз). Бонусы от программы лояльности и возвраты также зачисляются на кошелёк.",
			Keywords:  []string{"кошелёк", "пополнить", "баланс", "wallet", "как пополнить"},
			SortOrder: 1,
		},
		{
			Category:  domain.FAQCategoryWallet,
			Question:  "Как вывести деньги с кошелька?",
			Answer:    "Для владельцев: выплаты осуществляются на банковский счёт, указанный в платёжных реквизитах. Для клиентов: средства кошелька используются только для оплаты бронирований.",
			Keywords:  []string{"вывод", "вывести деньги", "выплата", "payout"},
			SortOrder: 2,
		},
		{
			Category:  domain.FAQCategoryAccount,
			Question:  "Как изменить номер телефона?",
			Answer:    "Перейдите в «Профиль» → «Настройки». Нажмите «Изменить» рядом с номером телефона и подтвердите новый номер кодом из SMS.",
			Keywords:  []string{"изменить телефон", "сменить номер", "новый телефон", "телефон"},
			SortOrder: 1,
		},
		{
			Category:  domain.FAQCategoryAccount,
			Question:  "Как удалить аккаунт?",
			Answer:    "Перейдите в «Профиль» → «Настройки» → «Удалить аккаунт». После подтверждения начнётся 30-дневный период, в течение которого вы можете передумать. По истечении срока данные будут безвозвратно удалены.",
			Keywords:  []string{"удалить аккаунт", "удалить профиль", "удаление", "delete account"},
			SortOrder: 2,
		},
		{
			Category:  domain.FAQCategoryGeneral,
			Question:  "Как связаться с поддержкой?",
			Answer:    "Создайте обращение через «Поддержка» в меню приложения. Опишите проблему, и наша команда ответит в ближайшее время. Среднее время ответа — менее 5 минут для срочных вопросов.",
			Keywords:  []string{"поддержка", "связаться", "помощь", "support", "контакт"},
			SortOrder: 1,
		},
		{
			Category:  domain.FAQCategoryGeneral,
			Question:  "Как оставить отзыв?",
			Answer:    "После завершения визита вы получите уведомление с предложением оставить отзыв. Также можно перейти в «Мои бронирования», выбрать завершённое и нажать «Оставить отзыв». Оцените по 4 критериям и добавьте фото/видео.",
			Keywords:  []string{"отзыв", "оценка", "review", "оставить отзыв", "написать отзыв"},
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
