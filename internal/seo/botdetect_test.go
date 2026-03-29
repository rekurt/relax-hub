package seo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBotUserAgent(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  bool
	}{
		{"empty", "", false},
		{"regular browser Chrome", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36", false},
		{"regular browser Firefox", "Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0", false},
		{"Googlebot", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", true},
		{"Googlebot smartphone", "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.216 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", true},
		{"YandexBot", "Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)", true},
		{"YandexMobileBot", "Mozilla/5.0 (iPhone; CPU iPhone OS 8_1 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Version/8.0 Mobile/12B411 Safari/600.1.4 (compatible; YandexMobileBot/3.0; +http://yandex.com/bots)", true},
		{"Bingbot", "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)", true},
		{"Facebook crawler", "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)", true},
		{"Twitter bot", "Twitterbot/1.0", true},
		{"Telegram bot", "TelegramBot (like TwitterBot)", true},
		{"VK share", "Mozilla/5.0 (compatible; vkShare; +http://vk.com/dev/Share)", true},
		{"LinkedIn bot", "LinkedInBot/1.0 (compatible; Mozilla/5.0; Apache-HttpClient +http://www.linkedin.com)", true},
		{"Slackbot", "Slackbot-LinkExpanding 1.0 (+https://api.slack.com/robots)", true},
		{"WhatsApp", "WhatsApp/2.23.20.0 A", true},
		{"Prerender", "Prerender (+https://github.com/prerender/prerender)", true},
		{"curl", "curl/8.1.2", false},
		{"wget", "Wget/1.21.4", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBotUserAgent(tt.userAgent)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSocialCrawler(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  bool
	}{
		{"empty", "", false},
		{"Googlebot is not social", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", false},
		{"Facebook is social", "facebookexternalhit/1.1", true},
		{"Twitter is social", "Twitterbot/1.0", true},
		{"VK is social", "Mozilla/5.0 (compatible; vkShare; +http://vk.com/dev/Share)", true},
		{"Telegram is social", "TelegramBot (like TwitterBot)", true},
		{"WhatsApp is social", "WhatsApp/2.23.20.0 A", true},
		{"Slack is social", "Slackbot-LinkExpanding 1.0", true},
		{"regular browser", "Mozilla/5.0 Chrome/120.0.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSocialCrawler(tt.userAgent)
			assert.Equal(t, tt.expected, result)
		})
	}
}
