package models

// Lazada
type Product struct {
	ItemID    	int      `json:"item_id"`
	Name        string 	 `json:"name"`
	Description string   `json:"description"`
	Images     	[]string `json:"images"` // List of product images
	Skus       	[]Sku    `json:"skus"`
	Quantity	int      `json:"quantity"`
}

type Sku struct {
	Status       string   `json:"status"`
	ShopSku      string   `json:"ShopSku"`
	Images       []string `json:"Images"` // List of SKU images
	Quantity     int      `json:"quantity"`
	Price        float64  `json:"price"`
	SpecialPrice float64  `json:"special_price"`
}

// API Response Struct
type ApiResponse struct {
	Code string `json:"code"`
	Data struct {
		Products []struct {
			ItemID      int             `json:"item_id"`
			Images      []string          `json:"images"`      // API sends as string (needs decoding)
			Skus        []Sku           `json:"skus"`
			Attributes  Attributes      `json:"attributes"`  // Separate struct for attributes
		} `json:"products"`
	} `json:"data"`
}

// Lazada Attributes Struct
type Attributes struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}