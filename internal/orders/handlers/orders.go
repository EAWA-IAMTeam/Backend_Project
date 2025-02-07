package handlers

import (
	"backend_project/internal/orders/services"
	"net/http"

	"github.com/labstack/echo/v4"
)

type OrdersHandler struct {
	service services.OrdersService
}

func NewOrdersHandler(service services.OrdersService) *OrdersHandler {
	return &OrdersHandler{service}
}

func (h *OrdersHandler) GetOrders(c echo.Context) error {
	createdAfter := c.QueryParam("created_after")
	if createdAfter == "" {
		createdAfter = "2025-02-01T22:44:30+08:00"
	}

	orders, err := h.service.GetOrders(createdAfter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Failed to retrieve orders",
			"error":   err.Error(),
		})
	}

	if len(orders) == 0 {
		return c.JSON(http.StatusOK, map[string]string{"message": "No orders found"})
	}

	return c.JSON(http.StatusOK, orders)
}

// func GetItemList(db *sql.DB, client *sdk.IopClient, appKey, accessToken string) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		orderIdParam := c.QueryParam("order_ids")
// 		if orderIdParam == "" {
// 			// Use some dummy IDs if none are provided
// 			orderIdParam = "465844362543475,463073081743475,465853375743475"
// 		}

// 		// Split the orderIdParam into a slice of strings
// 		orderIdStrings := strings.Split(orderIdParam, ",")

// 		// Convert the slice of strings to a slice of int64
// 		var orderIds []int64
// 		for _, idStr := range orderIdStrings {
// 			id, err := strconv.ParseInt(idStr, 10, 64)
// 			if err != nil {
// 				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid order ID format"})
// 			}
// 			orderIds = append(orderIds, id)
// 		}

// 		// Convert the slice of int64 to a comma-separated string for the API call
// 		orderIdsStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(orderIds)), ","), "[]")

// 		// Add square brackets to format as a JSON array
// 		orderIdsStr = "[" + orderIdsStr + "]"

// 		fmt.Println("orderIdsStr:", orderIdsStr)

// 		client.AddAPIParam("order_ids", orderIdsStr)

// 		queryParams := map[string]string{
// 			"appKey":      appKey,
// 			"accessToken": accessToken,
// 		}
// 		log.Println("Fetching order items for batch...")
// 		resp, err := client.Execute("/orders/items/get", "GET", queryParams)
// 		if err != nil {
// 			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to fetch order items", "error": err.Error()})
// 		}

// 		// Return the raw response to the client
// 		return c.JSON(http.StatusOK, resp.Data)
// 	}
// }
