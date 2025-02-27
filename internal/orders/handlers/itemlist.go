package handlers

import (
	"backend_project/internal/orders/services"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type ItemListHandler struct {
	service services.ItemListService
}

func NewItemListHandler(service services.ItemListService) *ItemListHandler {
	return &ItemListHandler{service}
}

func (h *ItemListHandler) GetItemList(c echo.Context) error {
	orderIDsParam := c.Param("order_ids")
	if orderIDsParam == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Order IDs are required"})
	}

	// Convert orderIDsParam to a slice of strings
	orderIDs := strings.Split(orderIDsParam, ",")

	items, err := h.service.GetItemList(orderIDs)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Failed to retrieve item list",
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, items)
}
