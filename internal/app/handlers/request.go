package handlers

// URLShortenRequest represents JSON {"url":"<some_url>"}
type URLShortenRequest struct {
	URL string `json:"url"`
}

type URLShortenCorrelatedRequest struct {
	CorrelatedID string `json:"correlated_id"`
	OriginalURL  string `json:"original_url"`
}

type BatchRequest []URLShortenCorrelatedRequest
