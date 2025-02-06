package models

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

type StockItem struct {
	ID               int64   `json:"stock_item_id"`
	CompanyID        int     `json:"company_id"`
	RefPrice         float64 `json:"ref_price"`
	RefCost          float64 `json:"ref_cost"`
	Quantity         int     `json:"quantity"`
	ReservedQuantity int     `json:"reserved_quantity"`
	StockCode        string  `json:"stock_code"`
	StockControl     bool    `json:"stock_control"`
	Weight           float64 `json:"weight"`
	Height           float64 `json:"height"`
	Width            float64 `json:"width"`
	Length           float64 `json:"length"`
	Variation1       string  `json:"variation1"`
	Variation2       string  `json:"variation2"`
	Platform         string  `json:"platform"`
	Description      string  `json:"description"`
	Status           bool    `json:"status"`
}

type StoreProduct struct {
	ID              int64   `json:"id"`
	StockItemID     int64   `json:"stock_item_id"`
	Price           float64 `json:"price"`
	DiscountedPrice float64 `json:"discounted_price"`
	SKU             string  `json:"sku"`
	Currency        string  `json:"currency"`
	Status          string  `json:"status"`
}

type MergeProduct struct {
	StockItemID   int64          `json:"stock_item_id"`
	RefPrice      float64        `json:"ref_price"`
	RefCost       float64        `json:"ref_cost"`
	Quantity      int            `json:"quantity"`
	StoreProducts []StoreProduct `json:"store_products"`
}

type ProductRequest struct {
	StoreID  int64          `json:"store_id"`
	Products []StoreProduct `json:"products" validate:"required,dive"`
}

type InsertResult struct {
	Inserted   int      `json:"inserted"`
	Duplicates []string `json:"duplicates"`
}
