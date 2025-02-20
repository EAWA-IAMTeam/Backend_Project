package handlers

import (
	"backend_project/internal/payment/models"
	"backend_project/internal/payment/services"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type PaymentHandler struct {
	paymentService services.PaymentService
}

// Constructor
func NewPaymentHandler(paymentService services.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// FUNCTION PART--------------------------------------------------------------

// Get Transactions
func (p *PaymentHandler) GetTransactions(c echo.Context) error {
	endTime := "2025-2-20T22:44:30+08:00"
	startTime := "2024-12-01T22:44:30+08:00"
	orderID := c.QueryParam("order_id")

	offset := 0
	limit := 500       // Maximum per request
	totalLimit := 5000 // Set a reasonable total limit to avoid infinite loops

	var allTransactions []models.LazadaTransaction

	for len(allTransactions) < totalLimit {
		transactions, err := p.paymentService.GetTransactions(startTime, endTime, orderID, offset, limit)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Failed to retrieve transactions",
				"error":   err.Error(),
			})
		}

		if len(transactions) == 0 {
			break // No more transactions to fetch
		}

		allTransactions = append(allTransactions, transactions...)
		offset += limit // Move to next batch

		// Ensure we don't exceed totalLimit
		if len(allTransactions) >= totalLimit {
			allTransactions = allTransactions[:totalLimit]
			break
		}
	}

	// Handle empty response
	if len(allTransactions) == 0 {
		return c.JSON(http.StatusOK, map[string]string{"message": "No transactions found"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"transactions": allTransactions})
}

// GetOrders
// func (p *PaymentHandler) GetOrders(c echo.Context) error {
// 	status := c.QueryParam("status")
// 	createdAfter := c.QueryParam("created_after")
// 	if createdAfter == "" {
// 		createdAfter = "2024-10-26T22:44:30+08:00"
// 	}
// 	// Fetch all orders in one call
// 	orders, err := p.paymentService.GetOrders(createdAfter, 0, 100, status)
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, map[string]string{
// 			"message": "Failed to retrieve orders",
// 			"error":   err.Error(),
// 		})
// 	}
// 	if orders == nil {
// 		log.Println("GetOrders() returned nil, initializing empty slice")
// 		orders = []models.LazadaOrder{}
// 	}
// 	if len(orders) == 0 {
// 		return c.JSON(http.StatusOK, map[string]string{"message": "No orders found"})
// 	}
// 	// Marshal orders into JSON for logging
// 	jsonOutput, err := json.MarshalIndent(orders, "", "  ")
// 	if err != nil {
// 		log.Println("Failed to marshal JSON:", err)
// 	} else {
// 		log.Println("Orders JSON Output:\n", string(jsonOutput))
// 	}
// 	return c.JSON(http.StatusOK, orders)
// }

// GetPayouts
func (p *PaymentHandler) GetPayouts(c echo.Context) error {
	createdAfter := c.QueryParam("created_after")
	if createdAfter == "" {
		createdAfter = "2024-10-01T22:44:30+08:00"
	}

	// Fetch all orders in one call
	payouts, err := p.paymentService.GetPayouts(createdAfter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Failed to retrieve orders",
			"error":   err.Error(),
		})
	}

	if payouts == nil {
		log.Println("GetOrders() returned nil, initializing empty slice")
		payouts = []models.LazadaPayout{}
	}

	if len(payouts) == 0 {
		return c.JSON(http.StatusOK, map[string]string{"message": "No orders found"})
	}

	return c.JSON(http.StatusOK, payouts)
}

// GetTransactionsByOrder
// func (p *PaymentHandler) GetTransactionsByOrder(c echo.Context) error {
// 	// Transaction testing param
// 	// endTime := time.Now().Format("2006-08-01T22:44:30+08:00")
// 	// startTime := time.Now().AddDate(0,0,-179)
// 	endTime := "2025-2-20T22:44:30+08:00"
// 	startTime := "2024-12-01T22:44:30+08:00"
// 	log.Print(startTime)
// 	// Order testing params
// 	createdAfter := c.QueryParam("created_after")
// 	if createdAfter == "" {
// 		createdAfter = startTime
// 	}
// 	orderOffset := 0
// 	orderLimit := 100
// 	transactionOffset := 0
// 	transactionLimit := 500
// 	status := c.QueryParam("status")
// 	// Get all orders first
// 	orders, err := p.paymentService.GetOrders(createdAfter, orderOffset, orderLimit, status)
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, map[string]string{
// 			"message": "Failed to retrieve orders",
// 			"error":   err.Error(),
// 		})
// 	}
// 	if len(orders) == 0 {
// 		return c.JSON(http.StatusOK, map[string]string{"message": "No orders found"})
// 	}
// 	// Initialize API response
// 	apiResponse := &models.LazadaAPIResponse{
// 		Order: orders,
// 	}
// 	// Store all transactions
// 	var allTransactions []models.LazadaTransaction
// 	// Fetch transactions for each order
// 	for _, order := range orders {
// 		// Convert id to string
// 		orderID := strconv.FormatInt(order.OrderNumber, 10)
// 		transactions, err := p.paymentService.GetTransactions(startTime, endTime, orderID, transactionOffset, transactionLimit)
// 		if err != nil {
// 			log.Printf("Error fetching transactions for OrderID %s: %v", orderID, err)
// 			continue
// 		}
// 		if transactions != nil {
// 			allTransactions = append(allTransactions, transactions...)
// 		}
// 	}
// 	// Assign fetched transactions to orders
// 	p.paymentService.AssignPaymentDataToOrder(apiResponse, allTransactions)
// 	return c.JSON(http.StatusOK, apiResponse)
// }

// GetTransactionsByPayout
func (p *PaymentHandler) GetTransactionsByPayout(c echo.Context) error {
	// Transaction testing param
	startTime := "2024-10-10T22:44:30+08:00"
	endTime := "2024-10-30T22:44:30+08:00"
	offset := 0
	limit := 500

	//Payout testing apram
	createdAfter := c.QueryParam("created_after")
	if createdAfter == "" {
		createdAfter = "2024-10-15T22:44:30+08:00"
	}

	// Get all payouts first
	payouts, err := p.paymentService.GetPayouts(createdAfter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Failed to retrieve orders",
			"error":   err.Error(),
		})
	}

	if len(payouts) == 0 {
		return c.JSON(http.StatusOK, map[string]string{"message": "No orders found"})
	}

	// Initialize API response
	apiResponse := &models.LazadaAPIResponse{
		Payout: payouts,
	}

	// Store all transactions
	var allTransactions []models.LazadaTransaction
	orderID := ""

	// Fetch transactions for each order
	transactions, err := p.paymentService.GetTransactions(startTime, endTime, orderID, offset, limit)
	log.Print(transactions)
	if err != nil {
		log.Print("Error fetching transactions", orderID, err)
	}

	//Append all transactions
	if transactions != nil {
		allTransactions = append(allTransactions, transactions...)
	}

	// Assign fetched transactions to orders
	p.paymentService.AssignPaymentDataToPayout(apiResponse, allTransactions)

	return c.JSON(http.StatusOK, apiResponse)
}
