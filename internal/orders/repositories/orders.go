package repositories

import (
	"backend_project/internal/orders/models"
	"backend_project/sdk"
	"encoding/json"
	"errors"
	"log"
)

type OrdersRepository interface {
	FetchOrders(createdAfter string) (*models.OrdersData, error)
}

type ordersRepository struct {
	client      *sdk.IopClient
	appKey      string
	accessToken string
}

func NewOrdersRepository(client *sdk.IopClient, appKey, accessToken string) OrdersRepository {
	return &ordersRepository{client, appKey, accessToken}
}

func (r *ordersRepository) FetchOrders(createdAfter string) (*models.OrdersData, error) {
	queryParams := map[string]string{
		"appKey":      r.appKey,
		"accessToken": r.accessToken,
	}

	r.client.AddAPIParam("created_after", createdAfter)

	resp, err := r.client.Execute("/orders/get", "GET", queryParams)
	if err != nil {
		log.Println("Error fetching orders:", err)
		return nil, err
	}

	log.Println("Raw orders response:", string(resp.Data))

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
