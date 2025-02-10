package models

type Item struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type OrderItem struct {
	ID    string `json:"id"`
	Items []Item `json:"items"`
}
