package handlers

import (
	"backend_project/internal/orders/models"
	"backend_project/internal/orders/services"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

type ProcessingState struct {
	LastProcessedTime string
	TotalProcessed    int
	IsCompleted       bool
	Offset            int
	CountTotal        int
	Orders            []models.Order
}

type OrderHandler struct {
	js              nats.JetStreamContext
	ordersService   services.OrdersService
	itemListService services.ItemListService
	returnHandler   *ReturnHandler
	paymentService  services.PaymentService
}

func NewOrdersHandler(js nats.JetStreamContext, ordersService services.OrdersService, itemListService services.ItemListService, returnHandler *ReturnHandler, paymentService services.PaymentService) *OrderHandler {
	return &OrderHandler{js: js, ordersService: ordersService, itemListService: itemListService, returnHandler: returnHandler, paymentService: paymentService}
}

// SetupSubscriptions initializes all NATS subscriptions
func (h *OrderHandler) SetupSubscriptions() error {
	// Subscribe to get orders by company
	if _, err := h.js.QueueSubscribe("order.request.getbycompany", "order-company", h.handleGetOrdersByCompany); err != nil {
		return err
	}
	if _, err := h.js.QueueSubscribe("order.request.getfromlazada", "order-lazada", h.GetOrders); err != nil {
		return err
	}
	// if _, err := h.js.QueueSubscribe("order.request.gettransactions", "order-lazada-transactions", h.GetTransactionsByOrder); err != nil {
	// 	return err
	// }

	log.Println("Order subscriptions setup complete")
	return nil
}

func (oh *OrderHandler) handleGetOrdersByCompany(msg *nats.Msg) {
	//Extract comapny ID from subject
	var request models.PaginatedRequest

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		log.Println("Failed to unmarshal request:", err)
		oh.respondWithError("Invalid request format", request.RequestID)
		msg.Ack()
		return
	}

	// Get orders from repository
	orders, _, err := oh.ordersService.FetchOrdersByCompanyID(
		request.CompanyID,
		request.Pagination.Page,
		request.Pagination.Limit,
	)
	if err != nil {
		log.Printf("Error fetching orders: %v", err)
		oh.respondWithError("Failed to fetch orders", request.RequestID)
		msg.Ack()
		return
	}

	// Send response using Jetstream
	response, err := json.Marshal(orders)
	if err != nil {
		oh.respondWithError("Internal server error", request.RequestID)
		msg.Ack()
		return
	}

	//Publish response to JetStream (`order.response.<requestID>`)
	responseSubject := "order.response." + request.RequestID
	if _, err := oh.js.Publish(responseSubject, response); err != nil {
		log.Printf("Error sending response: %v", err)
	}

	// Explicitly acknowledge message
	msg.Ack()
}

func (oh *OrderHandler) GetOrders(msg *nats.Msg) {
	//Extract comapny ID from subject
	var request models.PaginatedRequest

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		log.Println("Failed to unmarshal request:", err)
		oh.respondWithError("Invalid request format", request.RequestID)
		msg.Ack()
		return
	}
	// if request.CompanyID == 0 {
	// 	return c.JSON(http.StatusBadRequest, map[string]string{"message": "Company ID is required"})
	// }

	var status = request.Status

	// Parse createdAfter or set default
	var createdAfter = request.Created_after
	if createdAfter == "" {
		// Default to current time minus 3 months
		startTime := time.Now().AddDate(0, -3, 0)
		createdAfter = startTime.Format("2006-01-02T15:04:05-07:00")
	}

	// Parse start time
	startTime, err := time.Parse(time.RFC3339, createdAfter)
	if err != nil {
		oh.respondWithError("Invalid created_after date format", request.RequestID)
		log.Println("Invalid created_after date format:", err)
		msg.Ack()
	}

	// Get stop date from query param or default to 2020
	var stopAfter = request.Stop_after
	var stopTime time.Time
	if stopAfter != "" {
		var err error
		stopTime, err = time.Parse(time.RFC3339, stopAfter)
		if err != nil {
			oh.respondWithError("Invalid stop_after date format", request.RequestID)
			log.Println("Invalid stop_after date format:", err)
			msg.Ack()
		}
	} else {
		stopTime = time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	}

	// Set initial time window
	currentStartTime := startTime
	allOrders := make([]models.Order, 0)
	totalProcessedCount := 0

	for {
		// Set 3-month window
		currentEndTime := currentStartTime.AddDate(0, 0, 3)
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

		var sortDirection = request.Sort_direction
		if sortDirection == "" {
			sortDirection = "DESC"
		} else if sortDirection != "ASC" && sortDirection != "DESC" {
			oh.respondWithError("Sort_direction must be either ASC or DESC", request.RequestID)
			log.Println("Sort_direction must be either ASC or DESC:", err)
			msg.Ack()
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
				orders, count, err := oh.ordersService.GetOrders(
					state.LastProcessedTime,
					currentCreatedBefore,
					state.Offset,
					limit,
					status,
					sortDirection,
				)
				if err != nil {
					oh.respondWithError("Failed to retrieve orders", request.RequestID)
					log.Println("Failed to retrieve orders:", err)
					msg.Ack()
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

						err := oh.ordersService.SaveOrder(&order, request.CompanyID)
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

	response := map[string]interface{}{
		"total_processed": totalProcessedCount,
		"orders":          allOrders,
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		oh.respondWithError("Failed to marshal response", request.RequestID)
		log.Printf("Failed to marshal response: %v", err)
		msg.Ack()
		return
	}

	responseSubject := "order.response." + request.RequestID
	if _, err := oh.js.Publish(responseSubject, responseData); err != nil {
		log.Printf("Error sending response: %v", err)
		log.Println(responseSubject, response)
	}

	msg.Ack()
}

func (oh *OrderHandler) GetTransactionsByOrder(msg *nats.Msg) {
	var request models.PaginatedRequest

	// Parse createdAfter or set default
	createdAfter := request.Created_after
	if createdAfter == "" {
		startTime := time.Now().AddDate(0, -3, 0)
		createdAfter = startTime.Format("2006-01-02T15:04:05-07:00")
	}

	// Parse start time
	startTime, err := time.Parse(time.RFC3339, createdAfter)
	if err != nil {
		oh.respondWithError("Invalid created_after date format", request.RequestID)
		log.Println("Invalid created_after date format:", err)
		msg.Ack()
	}

	// Get stop date from query param or default to 2020
	var stopAfter = request.Stop_after
	var endTime time.Time
	if stopAfter != "" {
		endTime, err = time.Parse(time.RFC3339, stopAfter)
		if err != nil {
			oh.respondWithError("Invalid stop_after date format", request.RequestID)
			log.Println("Invalid stop_after date format:", err)
			msg.Ack()
		}
	} else {
		endTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	// Get all orders with embedded SQLData from DB
	companyID := request.CompanyID
	orders, _, err := oh.ordersService.FetchOrdersByCompanyID(companyID, request.Pagination.Limit, request.Pagination.Page)
	if err != nil {
		oh.respondWithError("Invalid Company ID", request.RequestID)
		log.Println("Invalid Company ID:", err)
		msg.Ack()
	}

	if len(orders) == 0 {
		oh.respondWithError("No orders found for this company", request.RequestID)
		log.Println("No orders found for this company:", err)
		msg.Ack()
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
			transactions, err := oh.paymentService.GetTransactions(currentStart, currentEnd, orderID, offset, limit)
			if err != nil {
				oh.respondWithError("Failed to retrieve transactions", request.RequestID)
				log.Println("Failed to retrieve transactions:", err)
				msg.Ack()
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
}

func (h *OrderHandler) respondWithError(errMsg string, requestID string) {
	response := map[string]string{"error": errMsg}
	data, _ := json.Marshal(response)

	responseSubject := "order.response." + requestID
	h.js.Publish(responseSubject, data)
}
