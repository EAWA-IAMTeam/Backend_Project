package models

import (
	"time"
)

// Order struct
type Order struct {
	OrderID         int64     `json:"order_id"`
	PlatformOrderID string    `json:"platform_order_id"`
	StoreID         string    `json:"store_id"`
	CompanyID       int64     `json:"company_id"`
	ShipmentDate    time.Time `json:"shipment_date"`
	OrderDate       time.Time `json:"order_date"`
	TrackingID      string    `json:"tracking_id"`
	OrderStatus     string    `json:"order_status"`
	Data            Data      `json:"data"`
	OrderItems      []Item    `json:"item_list"`
}

type Data struct {
	OrderID                   int64    `json:"order_id"`
	CustomerName              string   `json:"CustomerName"`
	CustomerPhone             string   `json:"CustomerPhone"`
	CustomerAddress           string   `json:"CustomerAddress"`
	CourierService            string   `json:"CourierService"`
	TransactionFee            float64  `json:"TransactionFee"`
	ShippingFee               float64  `json:"ShippingFee"`
	ProcessFee                float64  `json:"ProcessFee"`
	ServiceFee                float64  `json:"service_fee"`
	SellerDiscount            float64  `json:"seller_discount"`
	PlatformDiscount          float64  `json:"platform_discount"`
	ShippingFeeDiscountSeller float64  `json:"shipping_fee_discount_seller"`
	TotalPrice                string   `json:"TotalAmount"`
	Currency                  string   `json:"currency"`
	RefundAmount              int      `json:"refund_amount"`
	RefundReason              string   `json:"reason_text"`
	CreatedAt                 string   `json:"created_at"`
	SystemUpdateTime          string   `json:"updated_at"`
	Statuses                  []string `json:"statuses"`
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
}

type Address struct {
	Country   string `json:"country"`
	City      string `json:"city"`
	Address1  string `json:"address1"`
	PostCode  string `json:"post_code"`
	FirstName string `json:"first_name"`
	Phone     string `json:"phone"`
}
