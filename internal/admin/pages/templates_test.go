package pages

import (
	"html/template"
	"testing"
)

func TestStatusRu(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"pending", "Ожидает"},
		{"confirmed", "Подтверждено"},
		{"approved", "Одобрено"},
		{"rejected", "Отклонено"},
		{"cancelled", "Отменено"},
		{"completed", "Завершено"},
		{"hidden", "Скрыто"},
		{"unknown_status", "unknown_status"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := StatusRu(tt.input)
			if got != tt.want {
				t.Errorf("StatusRu(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestJsEscape(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain text", "hello world", "hello world"},
		{"double quote", `say "hello"`, `say \"hello\"`},
		{"single quote", "it's", `it\'s`},
		{"backslash", `path\to\file`, `path\\to\\file`},
		{"angle brackets", "<script>alert(1)</script>", `\x3cscript\x3ealert(1)\x3c/script\x3e`},
		{"ampersand", "foo&bar", `foo\x26bar`},
		{"newline", "line1\nline2", `line1\nline2`},
		{"carriage return", "line1\rline2", `line1\rline2`},
		{"russian text", "Русская баня на дровах", "Русская баня на дровах"},
		{"xss payload", `"; alert(document.cookie); "`, `\"; alert(document.cookie); \"`},
		{"empty", "", ""},
		{"combined", `Баня "Огонь & Пар" <новая>`, `Баня \"Огонь \x26 Пар\" \x3cновая\x3e`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JsEscape(tt.input)
			if got != template.JS(tt.want) {
				t.Errorf("JsEscape(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
