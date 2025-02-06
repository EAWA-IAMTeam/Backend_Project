package models

// Lazada
type Product struct {
	ItemID     int      `json:"item_id"`
	Images     []string `json:"images"` // List of product images
	Skus       []Sku    `json:"skus"`
	Attributes struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"attributes"`
}

type Sku struct {
	Status       string   `json:"status"`
	ShopSku      string   `json:"ShopSku"`
	Images       []string `json:"Images"` // List of SKU images
	Quantity     int      `json:"quantity"`
	Price        float64  `json:"price"`
	SpecialPrice float64  `json:"special_price"`
}

type ApiResponse struct {
	Code string `json:"code"`
	Data struct {
		Products []Product `json:"products"`
	} `json:"data"`
}