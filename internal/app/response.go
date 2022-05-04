package app

// URLShortenResponse represents JSON {"result":"<shorten_url>"}
type URLShortenResponse struct {
	Result string `json:"result"`
}
