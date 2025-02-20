package handlers

import (
	"backend_project/internal/orders/models"
	"backend_project/internal/orders/services"
	"fmt"
	"net/http"

	"log"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type OrdersHandler struct {
	ordersService   services.OrdersService
	itemListService services.ItemListService
	returnHandler   *ReturnHandler
}

func NewOrdersHandler(ordersService services.OrdersService, itemListService services.ItemListService, returnHandler *ReturnHandler) *OrdersHandler {
	return &OrdersHandler{ordersService, itemListService, returnHandler}
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

	// Set createdBefore to 3 months after createdAfter
	endTime := startTime.AddDate(0, 3, 0)
	createdBefore := endTime.Format("2006-01-02T15:04:05-07:00")

	// Override with query param if provided
	if queryBefore := c.QueryParam("created_before"); queryBefore != "" {
		parsedQueryBefore, err := time.Parse(time.RFC3339, queryBefore)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Invalid created_before date format",
				"error":   err.Error(),
			})
		}
		// Use provided date if it's within 3 months
		if parsedQueryBefore.Before(endTime) {
			createdBefore = queryBefore
		}
	}

	log.Printf("Processing orders from %s to %s", createdAfter, createdBefore)

	sortDirection := c.QueryParam("sort_direction")
	if sortDirection == "" {
		sortDirection = "DESC"
	} else if sortDirection != "ASC" && sortDirection != "DESC" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "sort_direction must be either ASC or DESC",
		})
	}

	// Processing state
	type ProcessingState struct {
		LastProcessedTime string
		TotalProcessed    int
		IsCompleted       bool
		Offset            int
		CountTotal        int
		Orders            []models.Order
	}

	state := ProcessingState{
		LastProcessedTime: createdAfter,
		TotalProcessed:    0,
		IsCompleted:       false,
		Offset:            0,
		CountTotal:        0,
		Orders:            make([]models.Order, 0),
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
				createdBefore,
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

	log.Printf("All orders processing completed. Total processed: %d", state.TotalProcessed)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"orders":          state.Orders,
		"total_processed": state.TotalProcessed,
		"count_total":     state.CountTotal,
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
