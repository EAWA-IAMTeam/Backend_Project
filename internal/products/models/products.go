package models

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
	ID              int64    `json:"id"`
	StoreID         string   `json:"store_id"`
	StockItemID     int64    `json:"stock_item_id"`
	Price           float64  `json:"price"`
	DiscountedPrice float64  `json:"discounted_price"`
	SKU             string   `json:"sku"`
	Currency        string   `json:"currency"`
	Status          string   `json:"status"`
	ImageURL        []string `json:"image_url"`
}

type MergeProduct struct {
	StockItemID   int64          `json:"stock_item_id"`
	RefPrice      float64        `json:"ref_price"`
	RefCost       float64        `json:"ref_cost"`
	Quantity      int            `json:"quantity"`
	StoreProducts []StoreProduct `json:"store_products"`
}

type InsertResult struct {
	Inserted   int      `json:"inserted"`
	Duplicates []string `json:"duplicates"`
}
