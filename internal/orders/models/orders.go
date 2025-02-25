package models

type OrdersData struct {
	CountTotal int     `json:"counttotal"`
	Count      int     `json:"count"`
	Orders     []Order `json:"orders"`
}

type Order struct {
	OrderID                     int64          `json:"order_id"`
	CreatedAt                   string         `json:"created_at"`
	UpdatedAt                   string         `json:"updated_at"`
	Price                       string         `json:"price"`
	PaymentMethod               string         `json:"payment_method"`
	Statuses                    []string       `json:"statuses"`
	VoucherPlatform             float64        `json:"voucher_platform"`
	Voucher                     float64        `json:"voucher"`
	WarehouseCode               string         `json:"warehouse_code"`
	VoucherSeller               float64        `json:"voucher_seller"`
	VoucherCode                 string         `json:"voucher_code"`
	GiftOption                  bool           `json:"gift_option"`
	ShippingFeeDiscountPlatform float64        `json:"shipping_fee_discount_platform"`
	CustomerLastName            string         `json:"customer_last_name"`
	PromisedShippingTimes       string         `json:"promised_shipping_times"`
	NationalRegistrationNumber  string         `json:"national_registration_number"`
	ShippingFeeOriginal         float64        `json:"shipping_fee_original"`
	BuyerNote                   string         `json:"buyer_note"`
	CustomerFirstName           string         `json:"customer_first_name"`
	ShippingFeeDiscountSeller   float64        `json:"shipping_fee_discount_seller"`
	ShippingFee                 float64        `json:"shipping_fee"`
	BranchNumber                string         `json:"branch_number"`
	TaxCode                     string         `json:"tax_code"`
	ItemsCount                  int            `json:"items_count"`
	DeliveryInfo                string         `json:"delivery_info"`
	ExtraAttributes             string         `json:"extra_attributes"`
	Remarks                     string         `json:"remarks"`
	GiftMessage                 string         `json:"gift_message"`
	TotalReleasedAmount         float64        `json:"total_released_amount"`
	AddressShipping             Address        `json:"address_shipping"`
	Items                       []Item         `json:"items"`
	RefundStatus                []ReturnRefund `json:"refund_status"`
}

type Address struct {
	Country   string `json:"country"`
	City      string `json:"city"`
	Address1  string `json:"address1"`
	PostCode  string `json:"post_code"`
	FirstName string `json:"first_name"`
	Phone     string `json:"phone"`
}

type SQLData struct {
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
	Status                    []string `json:"statuses"`
	//Platform string `json:"platform"` (set via prefix)
	//Store string `json:"store"` (set via prefix)
	//PlatformReleasedAmount float64 `json:"platform_released_amount"` //payment
	TotalReleasedAmount float64 `json:"total_released_amount"`
	PaymentMethod       string  `json:"payment_method"` //payment
	RefundAmount        int     `json:"refund_amount"`
	RefundReason        string  `json:"reason_text"`
	CreatedAt           string  `json:"created_at"`
	SystemUpdateTime    string  `json:"updated_at"`
}
