package models

// RequestPayload struct (for POST requests)
type RequestPayload struct {
	Topic string      `json:"topic"`
	Data  interface{} `json:"data"`
}
