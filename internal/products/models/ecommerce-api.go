package models

// Generalized Product Struct
type Product struct {
	ItemID    	int64    `json:"item_id"`
	StoreID   	int64    `json:"store_id"`
	Name        string 	 `json:"name"`
	Description string   `json:"description"`
	Images     	[]string `json:"images"`
	Skus       	[]Sku    `json:"skus"`
	Quantity	int      `json:"quantity"`
}

type Sku struct {
	Status       string   `json:"status"`
	ShopSku      string   `json:"ShopSku"`
	Images       []string `json:"Images"`
	Quantity     int      `json:"quantity"`
	Price        float64  `json:"price"`
	SpecialPrice float64  `json:"special_price"`
}

// Lazada API Response Product Struct
type ApiResponse struct {
	Code string `json:"code"`
	Data struct {
		Products []struct {
			ItemID      int64           `json:"item_id"`
			Images      []string        `json:"images"`      
			Skus        []Sku           `json:"skus"`
			Attributes  Attributes      `json:"attributes"` 
		} `json:"products"`
	} `json:"data"`
}

type Attributes struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}