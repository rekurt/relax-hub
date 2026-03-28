package payment

import (
	"errors"
	"net"
	"strings"
)

// PaymentErrorInfo holds classified payment error information.
type PaymentErrorInfo struct {
	Retryable  bool
	Code       string
	MessageRU  string
	Suggestion string
}

// ClassifyError determines whether a payment error is retryable and provides
// a user-friendly Russian message with a helpful suggestion.
func ClassifyError(err error) PaymentErrorInfo {
	if err == nil {
		return PaymentErrorInfo{}
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	// Network / timeout errors are retryable
	var netErr net.Error
	if errors.As(err, &netErr) {
		return PaymentErrorInfo{
			Retryable:  true,
			Code:       "network_error",
			MessageRU:  "Ошибка сети при обработке платежа",
			Suggestion: "Попробуйте ещё раз через несколько секунд",
		}
	}

	// Check for common retryable error patterns
	if containsAny(lower, "timeout", "timed out", "deadline exceeded") {
		return PaymentErrorInfo{
			Retryable:  true,
			Code:       "timeout",
			MessageRU:  "Превышено время ожидания ответа от платёжной системы",
			Suggestion: "Попробуйте ещё раз через несколько секунд",
		}
	}

	if containsAny(lower, "connection refused", "connection reset", "eof", "broken pipe") {
		return PaymentErrorInfo{
			Retryable:  true,
			Code:       "connection_error",
			MessageRU:  "Ошибка соединения с платёжной системой",
			Suggestion: "Попробуйте ещё раз через несколько секунд",
		}
	}

	if containsAny(lower, "too many requests", "rate limit", "429") {
		return PaymentErrorInfo{
			Retryable:  true,
			Code:       "rate_limited",
			MessageRU:  "Слишком много запросов к платёжной системе",
			Suggestion: "Подождите минуту и попробуйте снова",
		}
	}

	if containsAny(lower, "internal server error", "502", "503", "504", "service unavailable") {
		return PaymentErrorInfo{
			Retryable:  true,
			Code:       "provider_error",
			MessageRU:  "Временная ошибка платёжной системы",
			Suggestion: "Попробуйте ещё раз через несколько секунд",
		}
	}

	// Permanent errors — do not retry
	if containsAny(lower, "insufficient funds", "insufficient_funds", "недостаточно средств") {
		return PaymentErrorInfo{
			Retryable:  false,
			Code:       "insufficient_funds",
			MessageRU:  "Недостаточно средств на карте",
			Suggestion: "Проверьте баланс или попробуйте другую карту",
		}
	}

	if containsAny(lower, "card_declined", "declined", "отклонена") {
		return PaymentErrorInfo{
			Retryable:  false,
			Code:       "card_declined",
			MessageRU:  "Платёж отклонён банком",
			Suggestion: "Попробуйте другую карту или обратитесь в банк",
		}
	}

	if containsAny(lower, "expired_card", "expired card", "карта просрочена") {
		return PaymentErrorInfo{
			Retryable:  false,
			Code:       "expired_card",
			MessageRU:  "Срок действия карты истёк",
			Suggestion: "Используйте другую карту",
		}
	}

	if containsAny(lower, "invalid_card", "invalid card number", "неверный номер") {
		return PaymentErrorInfo{
			Retryable:  false,
			Code:       "invalid_card",
			MessageRU:  "Неверные данные карты",
			Suggestion: "Проверьте номер карты и попробуйте снова",
		}
	}

	if containsAny(lower, "3d_secure", "3ds", "authentication") {
		return PaymentErrorInfo{
			Retryable:  false,
			Code:       "3ds_failed",
			MessageRU:  "Не пройдена проверка 3D Secure",
			Suggestion: "Попробуйте ещё раз и подтвердите платёж в приложении банка",
		}
	}

	if containsAny(lower, "blocked", "заблокирована", "fraud") {
		return PaymentErrorInfo{
			Retryable:  false,
			Code:       "card_blocked",
			MessageRU:  "Карта заблокирована",
			Suggestion: "Обратитесь в банк или попробуйте другую карту",
		}
	}

	if containsAny(lower, "limit exceeded", "превышен лимит") {
		return PaymentErrorInfo{
			Retryable:  false,
			Code:       "limit_exceeded",
			MessageRU:  "Превышен лимит по карте",
			Suggestion: "Попробуйте другую карту или обратитесь в банк",
		}
	}

	// Unknown/unclassified error — not retryable by default
	return PaymentErrorInfo{
		Retryable:  false,
		Code:       "payment_error",
		MessageRU:  "Ошибка при обработке платежа",
		Suggestion: "Попробуйте другой способ оплаты или обратитесь в поддержку",
	}
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
