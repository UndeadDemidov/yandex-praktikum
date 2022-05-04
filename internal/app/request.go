package app

// URLShortenRequest represents JSON {"url":"<some_url>"}
type URLShortenRequest struct {
	URL string `json:"url"`
}
