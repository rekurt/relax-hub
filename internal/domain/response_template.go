package domain

import (
	"time"

	"github.com/google/uuid"
)

type ResponseTemplate struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	Title     string
	Body      string
	IsDefault bool
	SortOrder int
	CreatedAt time.Time
}

func (t *ResponseTemplate) Validate() error {
	if t.OwnerID == uuid.Nil {
		return ErrInvalidInput
	}
	if t.Title == "" || t.Body == "" {
		return ErrInvalidInput
	}
	if len(t.Title) > 200 {
		return ErrInvalidInput
	}
	if len(t.Body) > 2000 {
		return ErrInvalidInput
	}
	return nil
}

const MaxTemplatesPerOwner = 50

var DefaultTemplates = []struct {
	Title string
	Body  string
}{
	{
		Title: "Благодарность за отзыв",
		Body:  "Спасибо за ваш отзыв! Мы рады, что вам понравилось. Будем ждать вас снова!",
	},
	{
		Title: "Ответ на негативный отзыв",
		Body:  "Благодарим за обратную связь. Нам очень жаль, что ваш опыт не оправдал ожиданий. Мы обязательно учтём ваши замечания и постараемся стать лучше.",
	},
	{
		Title: "Приглашение вернуться",
		Body:  "Спасибо за отзыв! Будем рады видеть вас снова. В качестве извинения за неудобства предлагаем скидку на следующий визит.",
	},
}
