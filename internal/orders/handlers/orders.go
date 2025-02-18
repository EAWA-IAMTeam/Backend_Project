package handlers

import (
	"backend_project/internal/orders/models"
	"backend_project/internal/orders/services"
	"fmt"
	"net/http"

	"log"

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
	// No need to check if status is empty, as it's optional

	createdAfter := c.QueryParam("created_after")
	if createdAfter == "" {
		createdAfter = "2024-02-01T22:44:30+08:00"
	}

	var allOrders []models.Order
	offset := 0
	limit := 100      // API's maximum limit per call
	totalLimit := 100 // Your internal limit for the operation

	for len(allOrders) < totalLimit {
		orders, err := h.ordersService.GetOrders(createdAfter, offset, limit, status)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Failed to retrieve orders",
				"error":   err.Error(),
			})
		}

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
		return c.JSON(http.StatusOK, map[string]string{"message": "No orders found"})
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

	// Log each order ID with status "returned"
	for _, order := range allOrders {
		if order.Statuses[0] == "returned" {
			err := h.returnHandler.HandleReturnRequest(int(order.OrderID))
			if err != nil {
				log.Printf("Error processing return for order ID %d: %s", order.OrderID, err.Error())
			}
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

	return c.JSON(http.StatusOK, allOrders)
}
