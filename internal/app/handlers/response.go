package handlers

import "fmt"

// URLShortenResponse represents JSON {"result":"<shorten_url>"}
type URLShortenResponse struct {
	Result string `json:"result"`
}

// BucketItem представляет собой структуру, в которой требуется сериализовать список ссылок
//
//	[
//	  {
//	    "short_url": "https://...",
//	    "original_url": "https://..."
//	  }, ...
//	]
type BucketItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// MapToBucket создает корзину ссылок из `map[string]string`
func MapToBucket(baseURL string, m map[string]string) *[]BucketItem {
	bucket := make([]BucketItem, 0, len(m))
	for k, v := range m {
		bucket = append(bucket, BucketItem{
			ShortURL:    fmt.Sprintf("%s%s", baseURL, k),
			OriginalURL: v,
		})
	}
	return &bucket
}

// URLShortenCorrelatedResponse представляет собой структуру, в которой требуется сериализовать список ссылок
//
//	[
//	  {
//	    "correlation_id": "4444",
//	    "short_url": "https://..."
//	  }, ...
//	]
type URLShortenCorrelatedResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
