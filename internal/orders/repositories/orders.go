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
func (or *OrderRepository) GetOrdersByCompany(companyID int8, page, limit int) ([]*models.Order, int, error) {
	//Calculate offset
	offset := (page - 1) * limit

	// Get paginated data
	query := `
	SELECT id, platform_order_id, store_id, company_id, shipment_date, order_date, 
			tracking_id, status, data::json as data, item_list::json as item_list
	FROM "Order" 
	WHERE company_id = $1
	ORDER BY order_date DESC
	LIMIT $2 OFFSET $3`

	// query := `
	//     SELECT id, platform_order_id, store_id, company_id, shipment_date, order_date, tracking_id, status,
	//            data::json as data,
	//            item_list::json as item_list
	//     FROM "Order"
	//     WHERE company_id = $1`

	rows, err := or.DB.Query(query, companyID, limit, offset)
	if err != nil {
		return nil, 0, err
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
			return nil, 0, err
		}

		// // Log the raw JSON data
		// log.Printf("Raw data JSON: %s", string(dataJson))
		// log.Printf("Raw item list JSON: %s", string(OrderItems))

		err = json.Unmarshal(dataJson, &order.Data)
		if err != nil {
			return nil, 0, err
		}

		err = json.Unmarshal(OrderItems, &order.OrderItems)
		if err != nil {
			return nil, 0, err
		}

		orders = append(orders, &order)
	}

	return orders, 0, nil
}
