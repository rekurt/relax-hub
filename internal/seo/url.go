package seo

import "strings"

func normalizeAssetURL(baseURL, rawURL string) string {
	if rawURL == "" {
		return ""
	}
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		return rawURL
	}

	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		base = "https://bani.ru"
	}

	if strings.HasPrefix(rawURL, "/") {
		return base + rawURL
	}
	return base + "/" + rawURL
}

func normalizeAssetURLs(baseURL string, rawURLs []string) []string {
	urls := make([]string, 0, len(rawURLs))
	for _, rawURL := range rawURLs {
		urls = append(urls, normalizeAssetURL(baseURL, rawURL))
	}
	return urls
}
