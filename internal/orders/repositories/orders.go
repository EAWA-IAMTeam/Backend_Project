package repositories

import (
	"backend_project/internal/orders/models"
	"database/sql"
	"encoding/json"
)

type OrderRepository struct {
	DB *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{DB: db}
}

// GetOrdersByCompany fetches orders by company ID
func (or *OrderRepository) GetOrdersByCompany(companyID int8) ([]*models.Order, error) {
	query := `
        SELECT id, platform_order_id, store_id, company_id, shipment_date, order_date, tracking_id, status,
               data::json as data,
               item_list::json as item_list
        FROM "Order" 
        WHERE company_id = $1`

	rows, err := or.DB.Query(query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	var dataJson []byte
	var OrderItems []byte

	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.OrderID,
			&order.PlatformOrderID,
			&order.StoreID,
			&order.CompanyID,
			&order.ShipmentDate,
			&order.OrderDate,
			&order.TrackingID,
			&order.OrderStatus,
			&dataJson,
			&OrderItems,
		)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(dataJson, &order.Data)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(OrderItems, &order.OrderItems)
		if err != nil {
			return nil, err
		}

		orders = append(orders, &order)
	}

	return orders, nil
}
