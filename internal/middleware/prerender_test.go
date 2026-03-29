package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrerender_BotGetsRedirected(t *testing.T) {
	botHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html>prerendered</html>"))
	})

	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("SPA content"))
	})

	mw := Prerender(nil, botHandler)
	handler := mw(normalHandler)

	req := httptest.NewRequest("GET", "/bathhouses/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, "<html>prerendered</html>", rr.Body.String())
	assert.Equal(t, "text/html", rr.Header().Get("Content-Type"))
}

func TestPrerender_NormalUserPassesThrough(t *testing.T) {
	botHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html>prerendered</html>"))
	})

	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("SPA content"))
	})

	mw := Prerender(nil, botHandler)
	handler := mw(normalHandler)

	req := httptest.NewRequest("GET", "/bathhouses/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, "SPA content", rr.Body.String())
}

func TestPrerender_NilBotHandler(t *testing.T) {
	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("SPA content"))
	})

	mw := Prerender(nil, nil)
	handler := mw(normalHandler)

	req := httptest.NewRequest("GET", "/bathhouses/test", nil)
	req.Header.Set("User-Agent", "Googlebot/2.1")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Should fall through to normal handler when botHandler is nil
	assert.Equal(t, "SPA content", rr.Body.String())
}

func TestPrerender_EmptyUserAgent(t *testing.T) {
	botHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("prerendered"))
	})

	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("SPA content"))
	})

	mw := Prerender(nil, botHandler)
	handler := mw(normalHandler)

	req := httptest.NewRequest("GET", "/bathhouses/test", nil)
	// No User-Agent header
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, "SPA content", rr.Body.String())
}

func TestPrerender_YandexBot(t *testing.T) {
	botHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("prerendered"))
	})

	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("SPA"))
	})

	mw := Prerender(nil, botHandler)
	handler := mw(normalHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, "prerendered", rr.Body.String())
}

func TestPrerender_FacebookCrawler(t *testing.T) {
	botHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("prerendered"))
	})

	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("SPA"))
	})

	mw := Prerender(nil, botHandler)
	handler := mw(normalHandler)

	req := httptest.NewRequest("GET", "/moscow/banya-lyuks", nil)
	req.Header.Set("User-Agent", "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, "prerendered", rr.Body.String())
}
