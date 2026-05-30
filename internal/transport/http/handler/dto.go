package httphandler

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Alias    string `json:"alias"`
	ShortURL string `json:"short_url"`
}