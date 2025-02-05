package handlers

import (
	"backend_project/internal/orders/models"
	"database/sql"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

// GetOrders returns all orders from the database
func GetOrders(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query("SELECT * FROM orders")
		if err != nil {
			return c.JSON(500, map[string]string{"error": "Failed to get orders"})
		}
		defer rows.Close()

		orders := []models.Order{}
		for rows.Next() {

			order := models.Order{}
			err := rows.Scan(&order.ID, &order.Product.ID, &order.Quantity, &order.CustomerName)
			if err != nil {
				return c.JSON(500, map[string]string{"error": "Failed to get orders"})
			}

			orders = append(orders, order)
		}

		return c.JSON(200, orders)
	}

}
