package models

type OrderItem struct {
	OrderNumber int64  `json:"order_number"`
	OrderID     int64  `json:"order_id"`
	OrderItems  []Item `json:"order_items"`
}

type Item struct {
	OrderItemID               int64   `json:"order_item_id"`
	Name                      string  `json:"name"`
	Status                    string  `json:"status"`
	PaidPrice                 float64 `json:"paid_price"`
	ItemPrice                 float64 `json:"item_price"`
	Quantity                  int     `json:"quantity"`
	Sku                       string  `json:"sku"`
	ShopSku                   string  `json:"shop_sku"`
	TrackingCode              string  `json:"tracking_code"`
	ShippingProviderType      string  `json:"shipping_provider_type"`
	ShippingFeeOriginal       float64 `json:"shipping_fee_original"`
	ShippingFeeDiscountSeller float64 `json:"shipping_fee_discount_seller"`
	ShippingAmount            float64 `json:"shipping_amount"`
	OrderID                   int64   `json:"order_id"`
	ReturnStatus              string  `json:"return_status"`
	ReturnReason              string  `json:"reason"`
	ImageUrl                  string  `json:"product_main_image"`
	// Add other fields as necessary based on the JSON structure
}
