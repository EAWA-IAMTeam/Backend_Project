package models

type ReturnData struct {
	Total   int          `json:"total"`
	Success bool         `json:"success"`
	PageNo  int          `json:"page_no"`
	Items   []ReturnItem `json:"items"`
}

type ReturnItem struct {
	ReverseOrderLines []ReturnRefund `json:"reverse_order_lines"`
}

type ReturnRefund struct {
	Product             Product `json:"product"`
	TradeOrderCreated   int64   `json:"trade_order_gmt_create"`
	ReasonText          string  `json:"reason_text"`
	ItemUnitPrice       int     `json:"item_unit_price"`
	TradeOrderLineID    int64   `json:"trade_order_line_id"`
	ReturnOrderModified int64   `json:"return_order_line_gmt_modified"`
	RefundPaymentMethod string  `json:"refund_payment_method"`
	RefundAmount        int     `json:"refund_amount"`
}

type Product struct {
	ProductSKU string `json:"product_sku"`
}