package repositories

import (
	"backend_project/internal/orders/models"
	"backend_project/sdk"
	"encoding/json"
	"errors"
	"log"
	"strings"
)

type ItemListRepository interface {
	FetchItemList(orderIDs []string) ([]models.OrderItem, error)
}

type itemListRepository struct {
	client      *sdk.IopClient
	appKey      string
	accessToken string
}

func NewItemListRepository(client *sdk.IopClient, appKey, accessToken string) ItemListRepository {
	return &itemListRepository{client, appKey, accessToken}
}

func (r *itemListRepository) FetchItemList(orderIDs []string) ([]models.OrderItem, error) {
	if len(orderIDs) == 0 {
		return nil, errors.New("no order IDs provided")
	}

	queryParams := map[string]string{
		"appKey":      r.appKey,
		"accessToken": r.accessToken,
	}

	// Convert orderIDs slice to a comma-separated string and wrap with []
	orderIDParam := "[" + strings.Join(orderIDs, ",") + "]"

	r.client.AddAPIParam("order_ids", orderIDParam)

	resp, err := r.client.Execute("/orders/items/get", "GET", queryParams)
	if err != nil {
		return nil, err
	}

	// log.Printf("Raw response from API: %s", string(resp.Data))

	// Assuming the response is a JSON array of order items
	var orderItems []models.OrderItem
	err = json.Unmarshal(resp.Data, &orderItems)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		return nil, errors.New("failed to parse item list")
	}

	return orderItems, nil
}
