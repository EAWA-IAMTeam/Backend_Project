package models

// Store response from lazada api
type LazadaAPIResponse struct {
	Transaction []LazadaTransaction `json:"transactions"`
	Payout      []LazadaPayout      `json:"payouts"`
}

type LazadaTransaction struct {
	TransactionDate   string `json:"transaction_date"`
	OrderNo           string `json:"order_no"`
	Amount            string `json:"amount"`
	PaidStatus        string `json:"paid_status"`
	WHTAmount         string `json:"WHT_amount"`
	VATInAmount       string `json:"VAT_in_amount"`
	TransactionNumber string `json:"transaction_number"`
	Statement         string `json:"statement"`
}

type LazadaPayout struct {
	Subtotal1          string              `json:"subtotal1"`
	Subtotal2          string              `json:"subtotal2"`
	ShipmentFeeCredit  string              `json:"shipment_fee_credit"`
	Payout             string              `json:"payout"`
	ItemRevenue        string              `json:"item_revenue"`
	OtherRevenueTotal  string              `json:"other_revenue_total"`
	FeesTotal          string              `json:"fees_total"`
	Refunds            string              `json:"refunds"`
	GuaranteeDeposit   string              `json:"guarantee_deposit"`
	FeesOnRefundsTotal string              `json:"fees_on_refunds_total"`
	StatementNumber    string              `json:"statement_number"`
	ShipmentFee        string              `json:"shipment_fee"`
	CreatedAt          string              `json:"created_at"`
	UpdatedAt          string              `json:"updated_at"`
	Transactions       []LazadaTransaction `json:"transactions"`
}
