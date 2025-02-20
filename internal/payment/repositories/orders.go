package repositories

import (
	"backend_project/internal/payment/models"
	"backend_project/sdk"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
)

type OrdersRepository interface {
	FetchOrders(createdAfter string, createdBefore string, offset int, limit int, status string, sort_by string) (*models.OrdersData, error)
	SaveOrder(order *models.Order, companyID string) error
	FetchOrdersByCompanyID(companyID string) ([]models.Order, error)
}

type ordersRepository struct {
	client      *sdk.IopClient
	appKey      string
	accessToken string
	DB          *sql.DB
}

func NewOrdersRepository(client *sdk.IopClient, appKey, accessToken string, db *sql.DB) OrdersRepository {
	return &ordersRepository{client, appKey, accessToken, db}
}

// FUNCTION PART--------------------------------------------------------------

// FetchPayout through External API Lazada
func (r *ordersRepository) FetchOrders(createdAfter string, createdBefore string, offset int, limit int, status string, sort_direction string) (*models.OrdersData, error) {
	queryParams := map[string]string{
		"appKey":      r.appKey,
		"accessToken": r.accessToken,
	}

	if createdAfter != "" {
		r.client.AddAPIParam("created_after", createdAfter)
	}
	if createdBefore != "" {
		r.client.AddAPIParam("created_before", createdBefore)
	}
	r.client.AddAPIParam("offset", strconv.Itoa(offset))
	r.client.AddAPIParam("limit", strconv.Itoa(limit))
	if status != "" {
		r.client.AddAPIParam("status", status)
	}

	if sort_direction != "" {
		r.client.AddAPIParam("sort_direction", sort_direction)
	}

	resp, err := r.client.Execute("/orders/get", "GET", queryParams)
	if err != nil {
		log.Println("Error fetching orders:", err)
		return nil, err
	}

	var apiResponse models.OrdersData
	err = json.Unmarshal(resp.Data, &apiResponse)
	if err != nil {
		log.Println("JSON Unmarshal Error:", err)
		return nil, err
	}

	log.Printf("API response: %+v", apiResponse.CountTotal)

	if apiResponse.Orders == nil {
		log.Println("API response `data` field is missing or null")
		return nil, errors.New("no data returned from API")
	}

	return &apiResponse, nil
}

func (r *ordersRepository) SaveOrder(order *models.Order, companyID string) error {
	if len(order.Items) == 0 {
		return errors.New("order has no items")
	}

	// Check if the order already exists
	var exists bool
	err := r.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM \"Order\" WHERE platform_order_id = $1 AND company_id = $2)", order.OrderID, companyID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking order existence: %v", err)
		return err
	}

	itemListJSON, err := json.Marshal(order.Items)
	if err != nil {
		log.Printf("Error marshalling items: %v", err)
		return err
	}

	sqlData := ConvertOrderToSQLData(*order)
	sqlDataJSON, err := json.Marshal(sqlData)
	if err != nil {
		log.Printf("Error marshalling SQL data: %v", err)
		return err
	}

	if exists {
		// Update existing order
		query := `UPDATE "Order" SET store_id = $2, tracking_id = $3, status = $4, item_list = $5, data = $6, order_date = $7 WHERE platform_order_id = $1 AND company_id = $8`
		_, err = r.DB.Exec(query, order.OrderID, order.ItemsCount, order.Items[0].TrackingCode, order.Statuses[0], string(itemListJSON), string(sqlDataJSON), order.CreatedAt, companyID)
	} else {
		// Insert new order
		query := `INSERT INTO "Order" (platform_order_id, store_id, tracking_id, status, item_list, data, company_id, order_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
		_, err = r.DB.Exec(query, order.OrderID, order.ItemsCount, order.Items[0].TrackingCode, order.Statuses[0], string(itemListJSON), string(sqlDataJSON), companyID, order.CreatedAt)
	}

	if err != nil {
		log.Printf("Error saving order: %v", err)
		return err
	}

	return nil
}

func ConvertOrderToSQLData(order models.Order) models.SQLData {
	// Ensure there is at least one element in the RefundStatus slice before accessing
	var refundAmount float64
	var refundReason string

	if len(order.RefundStatus) > 0 {
		// Convert from ReturnRefund struct
		refundAmount = float64(order.RefundStatus[0].RefundAmount)
		refundReason = order.RefundStatus[0].ReasonText
	}

	return models.SQLData{
		OrderID:                   order.OrderID,
		CustomerName:              order.CustomerFirstName + " " + order.CustomerLastName,
		CustomerPhone:             order.AddressShipping.Phone,
		CustomerAddress:           order.AddressShipping.Address1,
		CourierService:            order.DeliveryInfo,
		TransactionFee:            0, // Assumption
		ShippingFee:               order.ShippingFee,
		ProcessFee:                0, // Assumption
		ServiceFee:                0, // Assumption
		SellerDiscount:            order.VoucherSeller,
		PlatformDiscount:          order.VoucherPlatform,
		ShippingFeeDiscountSeller: order.ShippingFeeDiscountSeller,
		TotalPrice:                order.Price,
		Currency:                  "MYR",
		TotalReleasedAmount:       order.TotalReleasedAmount,
		Status:                    order.Statuses,
		RefundAmount:              int(refundAmount),
		RefundReason:              refundReason,
		CreatedAt:                 order.CreatedAt,
		SystemUpdateTime:          order.UpdatedAt,
	}
}

func (r *ordersRepository) FetchOrdersByCompanyID(companyID string) ([]models.Order, error) {
	query := `
		SELECT platform_order_id, tracking_id, status, data, item_list
		FROM "Order"
		WHERE company_id = $1`

	rows, err := r.DB.Query(query, companyID)
	if err != nil {
		log.Printf("Error querying orders: %v", err)
		return nil, err
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
			return nil, err
		}

		// Parse the SQLData from JSON
		var sqlData models.SQLData
		if err := json.Unmarshal([]byte(dataJSON), &sqlData); err != nil {
			log.Printf("Error unmarshaling SQL data: %v", err)
			return nil, err
		}

		// Parse the ItemList from JSON
		if err := json.Unmarshal([]byte(itemListJSON), &order.Items); err != nil {
			log.Printf("Error unmarshaling item list: %v", err)
			return nil, err
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
		order.TotalReleasedAmount = sqlData.TotalReleasedAmount // Update with full status array from SQLData

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
		return nil, err
	}

	return orders, nil
}
