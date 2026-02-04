package sse

type SSEEvent struct {
	Type    string `json:"type"`
	ID      string `json:"id,omitempty"`
	Content string `json:"content,omitempty"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}
