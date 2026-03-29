package seo

import (
	"strings"
)

// Known bot user-agent substrings for search engines and social media crawlers.
var botUserAgents = []string{
	// Google
	"googlebot",
	"google-inspectiontool",
	"storebot-google",
	"googleother",
	"apis-google",
	"mediapartners-google",
	"adsbot-google",
	// Yandex
	"yandexbot",
	"yandexaccessibilitybot",
	"yandexmobilebot",
	"yandexdirectdyn",
	"yandexscreenshotbot",
	"yandeximages",
	"yandexvideo",
	"yandexmedia",
	"yandexblogs",
	"yandexpagechecker",
	"yandexwebmaster",
	"yandexnews",
	"yandexfavicons",
	"yandexmetrika",
	// Bing / Microsoft
	"bingbot",
	"msnbot",
	"bingpreview",
	// Social media
	"facebookexternalhit",
	"facebot",
	"twitterbot",
	"linkedinbot",
	"telegrambot",
	"whatsapp",
	"vkshare",
	"odnoklassnikibot",
	"pinterestbot",
	"slackbot",
	// Other search engines
	"duckduckbot",
	"baiduspider",
	"sogou",
	"ia_archiver",
	// SEO tools
	"semrushbot",
	"ahrefsbot",
	"dotbot",
	"rogerbot",
	"screaming frog",
	// Generic
	"bot",
	"spider",
	"crawler",
	"prerender",
}

// IsBotUserAgent returns true if the given User-Agent string matches a known
// search engine bot or social media crawler.
func IsBotUserAgent(userAgent string) bool {
	if userAgent == "" {
		return false
	}
	ua := strings.ToLower(userAgent)
	for _, bot := range botUserAgents {
		if strings.Contains(ua, bot) {
			return true
		}
	}
	return false
}

// IsSocialCrawler returns true if the User-Agent belongs to a social media
// crawler that needs OG meta tags (Facebook, Twitter, VK, etc.).
func IsSocialCrawler(userAgent string) bool {
	if userAgent == "" {
		return false
	}
	ua := strings.ToLower(userAgent)
	socialBots := []string{
		"facebookexternalhit", "facebot", "twitterbot",
		"linkedinbot", "telegrambot", "whatsapp",
		"vkshare", "odnoklassnikibot", "pinterestbot", "slackbot",
	}
	for _, bot := range socialBots {
		if strings.Contains(ua, bot) {
			return true
		}
	}
	return false
}
