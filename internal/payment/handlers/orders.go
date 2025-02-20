package handlers

import (
	"backend_project/internal/payment/models"
	"backend_project/internal/payment/services"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type OrdersHandler struct {
	ordersService   services.OrdersService
	itemListService services.ItemListService
	returnHandler   *ReturnHandler
	paymentService  services.PaymentService
}

func NewOrdersHandler(ordersService services.OrdersService, itemListService services.ItemListService, returnHandler *ReturnHandler, paymentService services.PaymentService) *OrdersHandler {
	return &OrdersHandler{ordersService, itemListService, returnHandler, paymentService}
}

func (h *OrdersHandler) GetOrders(c echo.Context) error {
	companyID := c.Param("company_id")
	if companyID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Company ID is required"})
	}

	status := c.Param("status")
	// No need to check if status is empty, as it's optional

	createdAfter := c.QueryParam("created_after")
	if createdAfter == "" {
		createdAfter = "2024-02-01T22:44:30+08:00"
	}
	createdBefore := c.QueryParam("created_before")

	sort_direction := c.QueryParam("sort_direction")
	if sort_direction == "" {
		sort_direction = "DESC" // Default to descending order
	} else if sort_direction != "ASC" && sort_direction != "DESC" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "sort_direction must be either ASC or DESC",
		})
	}

	var allOrders []models.Order
	offset := 0
	limit := 100       // API's maximum limit per call
	totalLimit := 100  // Your internal limit for the operation
	var totalCount int // Add at the top with other vars

	for len(allOrders) < totalLimit {
		orders, count, err := h.ordersService.GetOrders(createdAfter, createdBefore, offset, limit, status, sort_direction)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Failed to retrieve orders",
				"error":   err.Error(),
			})
		}

		totalCount = count // Store the count from the first response

		if len(orders) == 0 {
			break // No more orders to fetch
		}

		allOrders = append(allOrders, orders...)
		offset += limit // Move to the next set of orders

		if len(allOrders) >= totalLimit {
			allOrders = allOrders[:totalLimit] // Ensure not to exceed 500 orders
			break
		}
	}

	if len(allOrders) == 0 {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"orders":     []models.Order{},
			"counttotal": 0,
			"count":      0,
		})
	}

	// Save each order from the retrieved list, adding a placeholder for those with no items
	for _, order := range allOrders {
		if len(order.Items) == 0 {
			placeholderItem := models.Item{
				Name: "No Item",
				// Add other necessary fields with default values
			}
			order.Items = append(order.Items, placeholderItem)
		}

		err := h.ordersService.SaveOrder(&order, companyID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": fmt.Sprintf("Failed to save order with ID %d", order.OrderID),
				"error":   err.Error(),
			})
		}
	}
	// Extract order IDs
	var orderIDs []string
	for _, order := range allOrders {
		orderIDs = append(orderIDs, fmt.Sprintf("%d", order.OrderID))
	}

	// Fetch item lists in batches of 50 order IDs
	for i := 0; i < len(orderIDs); i += 50 {
		end := i + 50
		if end > len(orderIDs) {
			end = len(orderIDs)
		}
		batch := orderIDs[i:end]

		log.Printf("Processing batch of order IDs: %v", batch) // Debugging statement
		_, err := h.itemListService.GetItemList(batch)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Failed to retrieve item lists",
				"error":   err.Error(),
			})
		}
	}

	// Create response structure
	response := struct {
		CountTotal int            `json:"counttotal"`
		Count      int            `json:"count"`
		Orders     []models.Order `json:"orders"`
	}{
		CountTotal: totalCount,
		Count:      len(allOrders),
		Orders:     allOrders,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *OrdersHandler) FetchOrdersByCompanyID(c echo.Context) error {
	companyID := c.Param("company_id")
	orders, err := h.ordersService.FetchOrdersByCompanyID(companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, orders)
}

func (h *OrdersHandler) GetTransactionsByOrder(c echo.Context) error {
	// Define date range for transaction search
	endTime := "2025-02-20T22:44:30+08:00"
	startTime := "2024-12-01T22:44:30+08:00"
	offset := 0
	limit := 500
	totalLimit := 10000 // Prevent excessive API calls

	// Get all orders with embedded SQLData from DB
	companyID := c.Param("company_id")
	orders, err := h.ordersService.FetchOrdersByCompanyID(companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// If no orders found
	if len(orders) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No orders found for this company"})
	}

	// Store all transactions
	var allTransactions []models.LazadaTransaction
	orderID := "" // Empty orderID to fetch all transactions at once

	// Fetch all transactions first (pagination)
	for len(allTransactions) < totalLimit {
		transactions, err := h.paymentService.GetTransactions(startTime, endTime, orderID, offset, limit)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Failed to retrieve transactions",
				"error":   err.Error(),
			})
		}

		if len(transactions) == 0 {
			break // No more transactions
		}

		allTransactions = append(allTransactions, transactions...)
		offset += limit // Move to next batch

		// Ensure we don't exceed totalLimit
		if len(allTransactions) >= totalLimit {
			allTransactions = allTransactions[:totalLimit]
			break
		}
	}

	// Map to store total released amounts per OrderID
	paymentSumMap := make(map[int64]float64)

	// Process transactions and sum up payments per OrderID
	for _, transaction := range allTransactions {
		orderIDInt, err := strconv.ParseInt(transaction.OrderNo, 10, 64)
		if err != nil {
			log.Printf("Error converting orderID '%s' to integer: %v", transaction.OrderNo, err)
			continue
		}

		amountStr := strings.ReplaceAll(transaction.Amount, ",", "")
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			log.Printf("Error converting amount for Order %d: %v", orderIDInt, err)
			continue
		}

		paymentSumMap[orderIDInt] += amount
	}

	// Assign total payments to each order's SQLData
	for i := range orders {
		if totalPayment, exists := paymentSumMap[orders[i].OrderID]; exists {
			orders[i].TotalReleasedAmount = math.Round(totalPayment*100) / 100 // Round to 2 decimal places
		} else {
			orders[i].TotalReleasedAmount = 0.0
		}
	}

	// Return updated orders with calculated TotalReleasedAmount
	return c.JSON(http.StatusOK, orders)
}
