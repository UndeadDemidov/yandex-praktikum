package handlers

// URLShortenRequest represents JSON {"url":"<some_url>"}
type URLShortenRequest struct {
	URL string `json:"url"`
}

// URLShortenCorrelatedRequest представляет собой структуру, в которую требуется дериализовать список ссылок для сокращения
// [
//   {
//     "correlation_id": "4444",
//     "original_url": "https://..."
//   }, ...
// ]
type URLShortenCorrelatedRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}
