package handlers

import (
	"backend_project/internal/orders/models"
	"backend_project/internal/orders/services"
	"fmt"
	"net/http"

	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type ProcessingState struct {
	LastProcessedTime string
	TotalProcessed    int
	IsCompleted       bool
	Offset            int
	CountTotal        int
	Orders            []models.Order
}

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

	// Parse createdAfter or set default
	createdAfter := c.QueryParam("created_after")
	if createdAfter == "" {
		// Default to current time minus 3 months
		startTime := time.Now().AddDate(0, -3, 0)
		createdAfter = startTime.Format("2006-01-02T15:04:05-07:00")
	}

	// Parse start time
	startTime, err := time.Parse(time.RFC3339, createdAfter)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Invalid created_after date format",
			"error":   err.Error(),
		})
	}

	// Get stop date from query param or default to 2020
	stopAfter := c.QueryParam("stop_after")
	var stopTime time.Time
	if stopAfter != "" {
		var err error
		stopTime, err = time.Parse(time.RFC3339, stopAfter)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Invalid stop_after date format",
				"error":   err.Error(),
			})
		}
	} else {
		stopTime = time.Date(2024, 12, 12, 0, 0, 0, 0, time.UTC)
	}

	// Set initial time window
	currentStartTime := startTime
	allOrders := make([]models.Order, 0)
	totalProcessedCount := 0

	for {
		// Set 3-month window
		currentEndTime := currentStartTime.AddDate(0, 3, 0)
		currentCreatedAfter := currentStartTime.Format("2006-01-02T15:04:05-07:00")
		currentCreatedBefore := currentEndTime.Format("2006-01-02T15:04:05-07:00")

		log.Printf("Processing orders from %s to %s", currentCreatedAfter, currentCreatedBefore)

		// Process current window
		state := ProcessingState{
			LastProcessedTime: currentCreatedAfter,
			TotalProcessed:    0,
			IsCompleted:       false,
			Offset:            0,
			CountTotal:        0,
			Orders:            make([]models.Order, 0),
		}

		sortDirection := c.QueryParam("sort_direction")
		if sortDirection == "" {
			sortDirection = "DESC"
		} else if sortDirection != "ASC" && sortDirection != "DESC" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "sort_direction must be either ASC or DESC",
			})
		}

		for !state.IsCompleted {
			var batchOrders []models.Order
			limit := 100
			maxOrdersPerBatch := 4900 // Batch processing limit

			// Error handling and concurrency control
			errorChan := make(chan error, maxOrdersPerBatch)
			sem := make(chan struct{}, 50)
			var wg sync.WaitGroup

			log.Printf("Starting batch processing from timestamp: %s with offset %d", state.LastProcessedTime, state.Offset)

			// Fetch orders in a single batch up to maxOrdersPerBatch
			batchCount := 0
			for batchCount < maxOrdersPerBatch {
				orders, count, err := h.ordersService.GetOrders(
					state.LastProcessedTime,
					currentCreatedBefore,
					state.Offset,
					limit,
					status,
					sortDirection,
				)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{
						"message": "Failed to retrieve orders",
						"error":   err.Error(),
					})
				}

				// Set total count from API response if it's the first batch
				if state.CountTotal == 0 {
					state.CountTotal = count
					log.Printf("Total available orders: %d", state.CountTotal)
				}

				log.Printf("Retrieved %d orders out of %d total", len(orders), count)

				if len(orders) == 0 || state.TotalProcessed >= state.CountTotal {
					state.IsCompleted = true
					break
				}

				// Save orders concurrently
				for _, order := range orders {
					wg.Add(1)
					sem <- struct{}{}
					go func(order models.Order) {
						defer wg.Done()
						defer func() { <-sem }() // Release semaphore

						if len(order.Items) == 0 {
							placeholderItem := models.Item{
								Name: "No Item",
							}
							order.Items = append(order.Items, placeholderItem)
						}

						err := h.ordersService.SaveOrder(&order, companyID)
						if err != nil {
							errorChan <- fmt.Errorf("failed to save order %d: %v", order.OrderID, err)
						}
					}(order)
				}

				batchOrders = append(batchOrders, orders...)
				state.Orders = append(state.Orders, orders...)
				state.Offset += limit
				state.TotalProcessed += len(orders)
				batchCount += len(orders)

				// Stop processing if we reached the total count
				if state.TotalProcessed >= state.CountTotal {
					state.IsCompleted = true
					break
				}

				// Stop if max batch size is reached
				if batchCount >= maxOrdersPerBatch {
					break
				}
			}

			// Wait for all saves in this batch to complete
			wg.Wait()
			close(errorChan)

			// Handle errors
			var errors []string
			for err := range errorChan {
				errors = append(errors, err.Error())
			}
			if len(errors) > 0 {
				log.Printf("Errors during order saving: %v", errors)
			}

			// Update state
			if len(batchOrders) > 0 {
				// Get the last order's timestamp and format to ISO8601
				lastOrder := batchOrders[len(batchOrders)-1]
				state.LastProcessedTime = lastOrder.CreatedAt

				// Optional: Add a small delay between batches
				time.Sleep(1 * time.Second)
			} else {
				state.IsCompleted = true
			}

			// Final check to ensure we don't exceed countTotal
			if state.TotalProcessed >= state.CountTotal {
				log.Printf("Reached total order count (%d), stopping process.", state.CountTotal)
				state.IsCompleted = true
			}
		}

		// Collect orders from this window
		allOrders = append(allOrders, state.Orders...)
		totalProcessedCount += state.TotalProcessed

		// Move window back by 3 months
		currentStartTime = currentStartTime.AddDate(0, -3, 0)
		if currentStartTime.Before(stopTime) {
			break
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"orders":          allOrders,
		"total_processed": totalProcessedCount,
	})
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
	// Parse createdAfter or set default
	createdAfter := c.QueryParam("created_after")
	if createdAfter == "" {
		startTime := time.Now().AddDate(0, -3, 0)
		createdAfter = startTime.Format("2006-01-02T15:04:05-07:00")
	}

	// Parse start time
	startTime, err := time.Parse(time.RFC3339, createdAfter)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Invalid created_after date format",
			"error":   err.Error(),
		})
	}

	// Get stop date from query param or default to 2020
	stopAfter := c.QueryParam("stop_after")
	var endTime time.Time
	if stopAfter != "" {
		endTime, err = time.Parse(time.RFC3339, stopAfter)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Invalid stop_after date format",
				"error":   err.Error(),
			})
		}
	} else {
		endTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	// Get all orders with embedded SQLData from DB
	companyID := c.Param("company_id")
	orders, err := h.ordersService.FetchOrdersByCompanyID(companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if len(orders) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No orders found for this company"})
	}

	// Store all transactions
	var allTransactions []models.LazadaTransaction
	orderID := ""
	offset := 0
	limit := 500
	totalLimit := 10000

	// Process transactions in 3-month windows
	currentStartTime := startTime
	for {
		currentEndTime := currentStartTime.AddDate(0, 3, 0)
		currentStart := currentStartTime.Format("2006-01-02T15:04:05-07:00")
		currentEnd := currentEndTime.Format("2006-01-02T15:04:05-07:00")

		log.Printf("Fetching transactions from %s to %s", currentStart, currentEnd)

		// Fetch transactions for current window
		for len(allTransactions) < totalLimit {
			transactions, err := h.paymentService.GetTransactions(currentStart, currentEnd, orderID, offset, limit)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"message": "Failed to retrieve transactions",
					"error":   err.Error(),
				})
			}

			if len(transactions) == 0 {
				break
			}

			allTransactions = append(allTransactions, transactions...)
			offset += limit

			if len(allTransactions) >= totalLimit {
				allTransactions = allTransactions[:totalLimit]
				break
			}
		}

		// Move window back by 3 months
		currentStartTime = currentStartTime.AddDate(0, -3, 0)
		if currentStartTime.Before(endTime) {
			break
		}
	}

	// Process transactions and update orders
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
