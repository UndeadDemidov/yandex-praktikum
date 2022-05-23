package handlers

import "fmt"

// URLShortenResponse represents JSON {"result":"<shorten_url>"}
type URLShortenResponse struct {
	Result string `json:"result"`
}

// BucketItem представляет собой структуру, в которой требуется сериализовать список ссылок
// [
//   {
//     "short_url": "https://...",
//     "original_url": "https://..."
//   }, ...
// ]
type BucketItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

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

type URLShortenCorrelatedResponse struct {
	CorrelatedID string `json:"correlated_id"`
	OriginalURL  string `json:"original_url"`
}

type BatchResponse []URLShortenCorrelatedRequest
