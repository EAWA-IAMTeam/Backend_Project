package handlers

import (
	"backend_project/internal/orders/services"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type OrdersHandler struct {
	ordersService   services.OrdersService
	itemListService services.ItemListService
}

func NewOrdersHandler(ordersService services.OrdersService, itemListService services.ItemListService) *OrdersHandler {
	return &OrdersHandler{ordersService, itemListService}
}

func (h *OrdersHandler) GetOrders(c echo.Context) error {
	createdAfter := c.QueryParam("created_after")
	if createdAfter == "" {
		createdAfter = "2025-02-01T22:44:30+08:00"
	}

	orders, err := h.ordersService.GetOrders(createdAfter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Failed to retrieve orders",
			"error":   err.Error(),
		})
	}

	if len(orders) == 0 {
		return c.JSON(http.StatusOK, map[string]string{"message": "No orders found"})
	}

	// Extract order IDs
	var orderIDs []string
	for _, order := range orders {
		orderIDs = append(orderIDs, fmt.Sprintf("%d", order.OrderID))
	}

	// Fetch item lists in batches of 50 order IDs
	for i := 0; i < len(orderIDs); i += 50 {
		end := i + 50
		if end > len(orderIDs) {
			end = len(orderIDs)
		}
		batch := orderIDs[i:end]

		_, err := h.itemListService.GetItemList(batch)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Failed to retrieve item lists",
				"error":   err.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, orders)
}
