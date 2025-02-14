package repositories

import (
	"backend_project/internal/orders/models"
	"backend_project/sdk"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strconv"
)

type OrdersRepository interface {
	FetchOrders(createdAfter string, offset int, limit int, status string) (*models.OrdersData, error)
	SaveOrder(order *models.Order, companyID string) error
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

func (r *ordersRepository) FetchOrders(createdAfter string, offset int, limit int, status string) (*models.OrdersData, error) {
	queryParams := map[string]string{
		"appKey":      r.appKey,
		"accessToken": r.accessToken,
	}

	r.client.AddAPIParam("created_after", createdAfter)
	r.client.AddAPIParam("offset", strconv.Itoa(offset))
	r.client.AddAPIParam("limit", strconv.Itoa(limit))
	if status != "" {
		r.client.AddAPIParam("status", status)
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

	// Check if the order already exists for the given company
	var exists bool
	err := r.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM \"Order\" WHERE id = $1 AND company_id = $2)", order.OrderID, companyID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking order existence: %v", err)
		return err
	}

	if exists {
		log.Printf("Order with ID %d already exists for company ID %s, skipping save", order.OrderID, companyID)
		return nil
	}

	query := `
		INSERT INTO "Order" (id, store_id, tracking_id, status, item_list, data, company_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	itemListJSON, err := json.Marshal(order.Items)
	if err != nil {
		log.Printf("Error marshalling items: %v", err)
		return err
	}

	// Serialize the entire order to JSON
	orderJSON, err := json.Marshal(order)
	if err != nil {
		log.Printf("Error marshalling order: %v", err)
		return err
	}

	_, err = r.DB.Exec(query, order.OrderID, order.ItemsCount, order.Items[0].TrackingCode, order.Statuses[0], string(itemListJSON), string(orderJSON), companyID)
	if err != nil {
		log.Printf("Error saving order: %v", err)
		return err
	}
	return nil
}
