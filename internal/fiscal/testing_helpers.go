package fiscal

// SetBaseURL overrides the base URL for testing purposes.
func (p *ATOLProvider) SetBaseURL(url string) {
	p.baseURL = url
}
