package repositories

import (
	"backend_project/internal/orders/models"
	"database/sql"
	"encoding/json"
	"log"
	"strings"
)

type OrderRepository struct {
	DB *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{DB: db}
}

func (r *OrderRepository) FetchOrdersByCompanyID(companyID int64, page, limit int) ([]models.Order, int, error) {
	//Calculate offset
	offset := (page - 1) * limit
	query := `
		SELECT platform_order_id, tracking_id, status, data, item_list
		FROM "Order"
		WHERE company_id = $1
		ORDER BY order_date DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.DB.Query(query, companyID, limit, offset)
	log.Print(rows)
	if err != nil {
		log.Printf("Error querying orders: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		// Initialize order with empty slices to prevent nil pointer dereference
		order := models.Order{
			Items:           make([]models.Item, 1), // Initialize with length 1 for first item
			Statuses:        make([]string, 1),      // Initialize with length 1 for first status
			AddressShipping: models.Address{},       // Initialize empty address struct
		}

		var dataJSON, itemListJSON string

		err := rows.Scan(
			&order.OrderID,
			&order.Items[0].TrackingCode,
			&order.Statuses[0],
			&dataJSON,
			&itemListJSON,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, 0, err
		}

		// Parse the SQLData from JSON
		var sqlData models.Data
		if err := json.Unmarshal([]byte(dataJSON), &sqlData); err != nil {
			log.Printf("Error unmarshaling SQL data: %v", err)
			return nil, 0, err
		}

		// Parse the ItemList from JSON
		if err := json.Unmarshal([]byte(itemListJSON), &order.Items); err != nil {
			log.Printf("Error unmarshaling item list: %v", err)
			return nil, 0, err
		}

		// Split customer name into first and last name
		names := strings.Split(sqlData.CustomerName, " ")
		if len(names) > 0 {
			order.CustomerFirstName = names[0]
			if len(names) > 1 {
				order.CustomerLastName = strings.Join(names[1:], " ")
			}
		}

		// Populate the order struct with data from SQLData
		order.AddressShipping.Phone = sqlData.CustomerPhone
		order.AddressShipping.FirstName = sqlData.CustomerName
		order.AddressShipping.Address1 = sqlData.CustomerAddress
		order.DeliveryInfo = sqlData.CourierService
		order.ShippingFee = sqlData.ShippingFee
		order.VoucherSeller = sqlData.SellerDiscount
		order.VoucherPlatform = sqlData.PlatformDiscount
		order.ShippingFeeDiscountSeller = sqlData.ShippingFeeDiscountSeller
		order.Price = sqlData.TotalPrice
		order.CreatedAt = sqlData.CreatedAt
		order.UpdatedAt = sqlData.SystemUpdateTime
		order.Statuses = sqlData.Status
		order.TotalReleasedAmount = sqlData.TotalReleasedAmount
		order.PaymentMethod = sqlData.PaymentMethod // Update with full status array from SQLData

		if sqlData.RefundAmount > 0 {
			order.RefundStatus = []models.ReturnRefund{{
				RefundAmount: sqlData.RefundAmount,
				ReasonText:   sqlData.RefundReason,
			}}
		}

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, 0, err
	}

	return orders, 0, nil
}

// // GetOrdersByCompany fetches orders by company ID
// func (or *OrderRepository) GetOrdersByCompany(companyID int8, page, limit int) ([]*models.Order, int, error) {
// 	//Calculate offset
// 	offset := (page - 1) * limit

// 	// Get paginated data
// 	query := `
// 	SELECT platform_order_id, tracking_id, status, data::json as data, item_list::json as item_list
// 	FROM "Order"
// 	WHERE company_id = $1
// 	ORDER BY order_date DESC
// 	LIMIT $2 OFFSET $3`

// 	// query := `
// 	//     SELECT id, platform_order_id, store_id, company_id, shipment_date, order_date, tracking_id, status,
// 	//            data::json as data,
// 	//            item_list::json as item_list
// 	//     FROM "Order"
// 	//     WHERE company_id = $1`

// 	rows, err := or.DB.Query(query, companyID, limit, offset)
// 	if err != nil {
// 		return nil, 0, err
// 	}
// 	defer rows.Close()

// 	var orders []*models.Order
// 	var dataJson []byte
// 	var OrderItems []byte

// 	for rows.Next() {
// 		var order models.Order
// 		err := rows.Scan(
// 			&order.OrderID,
// 			&order.PlatformOrderID,
// 			&order.StoreID,
// 			&order.CompanyID,
// 			&order.ShipmentDate,
// 			&order.OrderDate,
// 			&order.Tr
// 			&order.OrderStatus,
// 			&dataJson,
// 			&OrderItems,
// 		)
// 		if err != nil {
// 			return nil, 0, err
// 		}

// 		// // Log the raw JSON data
// 		// log.Printf("Raw data JSON: %s", string(dataJson))
// 		// log.Printf("Raw item list JSON: %s", string(OrderItems))

// 		err = json.Unmarshal(dataJson, &order.Data)
// 		if err != nil {
// 			return nil, 0, err
// 		}

// 		err = json.Unmarshal(OrderItems, &order.OrderItems)
// 		if err != nil {
// 			return nil, 0, err
// 		}

// 		orders = append(orders, &order)
// 	}

// 	return orders, 0, nil
// }
