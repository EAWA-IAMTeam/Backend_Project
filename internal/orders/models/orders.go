package models

type OrdersData struct {
	Orders []Order `json:"orders"`
}

type Order struct {
	OrderNumber                 int64    `json:"order_number"`
	CreatedAt                   string   `json:"created_at"`
	UpdatedAt                   string   `json:"updated_at"`
	Price                       string   `json:"price"`
	PaymentMethod               string   `json:"payment_method"`
	Statuses                    []string `json:"statuses"`
	OrderID                     int64    `json:"order_id"`
	VoucherPlatform             float64  `json:"voucher_platform"`
	Voucher                     float64  `json:"voucher"`
	WarehouseCode               string   `json:"warehouse_code"`
	VoucherSeller               float64  `json:"voucher_seller"`
	VoucherCode                 string   `json:"voucher_code"`
	GiftOption                  bool     `json:"gift_option"`
	ShippingFeeDiscountPlatform float64  `json:"shipping_fee_discount_platform"`
	CustomerLastName            string   `json:"customer_last_name"`
	PromisedShippingTimes       string   `json:"promised_shipping_times"`
	NationalRegistrationNumber  string   `json:"national_registration_number"`
	ShippingFeeOriginal         float64  `json:"shipping_fee_original"`
	BuyerNote                   string   `json:"buyer_note"`
	CustomerFirstName           string   `json:"customer_first_name"`
	ShippingFeeDiscountSeller   float64  `json:"shipping_fee_discount_seller"`
	ShippingFee                 float64  `json:"shipping_fee"`
	BranchNumber                string   `json:"branch_number"`
	TaxCode                     string   `json:"tax_code"`
	ItemsCount                  int      `json:"items_count"`
	DeliveryInfo                string   `json:"delivery_info"`
	ExtraAttributes             string   `json:"extra_attributes"`
	Remarks                     string   `json:"remarks"`
	GiftMessage                 string   `json:"gift_message"`
	AddressShipping             Address  `json:"address_shipping"`
	Items                       []Item   `json:"items"`
}

// platform_discount (order)
// seller_discount (order)
// refund_amount
// cancel_reason
// shipping_fee
// total_amount
// currency
// payment_type
// cust_name
// cust_phone
// tracking_number laz(GetOrderTrace) shopee
// shipment_date (created_at)
// shipment_courier  Laz(OrderDetails)Shopee(OrderDetails)
// recipient_address (order)
type Address struct {
	Country   string `json:"country"`
	City      string `json:"city"`
	Address1  string `json:"address1"`
	PostCode  string `json:"post_code"`
	FirstName string `json:"first_name"`
	Phone     string `json:"phone"`
}

// type OrderData struct {
// 	CustomerName string `json:"first_name"`
// 	CustomerPhone string `json:"phone"`
// 	CustomerAddress string `json:"address1"`
// 	CourierService string `json:"shipping_provider_type"`
// 	TransactionFee float64 `json:"transaction_fee"`
// 	ShippingFee float64 `json:"shipping_fee"`
// 	ProcessFee float64 `json:"process_fee"`
// 	ServiceFee float64 `json:"service_fee"`
// 	SellerDiscount float64 `json:"seller_discount"`
// 	PlatformDiscount float64 `json:"platform_discount"`
// 	ShippingFeeDiscountSeller float64 `json:"shipping_fee_discount_seller"`
// 	TotalPrice float64 `json:"price"`
// 	Currency string `json:"currency"`
// 	//Platform string `json:"platform"` (set via prefix)
// 	//Store string `json:"store"` (set via prefix)
// 	//PlatformReleasedAmount float64 `json:"platform_released_amount"` //payment
// 	//TotalReleasedAmount float64 `json:"total_released_amount"` //payment
// 	RefundAmount float64 `json:"refund_amount"`

// }
