package repositories

import (
	"backend_project/internal/payment/models"
	"backend_project/sdk"
	"database/sql"
	"encoding/json"
	"log"
	"strconv"
)

type PaymentRepository interface {
    FetchTransactions(startTime string, endTime string, orderID string, offset int, limit int) (*models.LazadaAPIResponse, error)
    FetchPayouts(createdAfter string) (*models.LazadaAPIResponse, error)
}

type paymentsRepository struct {
	client      *sdk.IopClient
	appKey      string
	accessToken string
	DB          *sql.DB
}

func NewPaymentsRepository(client *sdk.IopClient, appKey, accessToken string, db *sql.DB) PaymentRepository {
	return &paymentsRepository{client, appKey, accessToken, db}
}

// FUNCTION PART--------------------------------------------------------------

//FetchPayout through External API Lazada
func (r *paymentsRepository) FetchTransactions(startTime string, endTime string, order_id string, offset int, limit int) (*models.LazadaAPIResponse, error) {
	queryParams := make(map[string]string)

	r.client.AddAPIParam("start_time", startTime)
	r.client.AddAPIParam("end_time", endTime)
	if order_id != "" {
		r.client.AddAPIParam("trade_order_id", order_id)
	}
	r.client.AddAPIParam("offset", strconv.Itoa(offset))
	r.client.AddAPIParam("limit", strconv.Itoa(limit))

	resp, err := r.client.Execute("/finance/transaction/details/get", "GET", queryParams)
	if err != nil {
		log.Println("Error fetching transactions:", err)
		return nil, err
	}

	var apiResponse models.LazadaAPIResponse
	err = json.Unmarshal(resp.Data, &apiResponse.Transaction)
	if err != nil {
		log.Println("JSON Unmarshal Error:", err)
		return nil, err
	}

	if apiResponse.Transaction == nil {
		log.Println("API response `data` field is missing or null")
		return &models.LazadaAPIResponse{Transaction: []models.LazadaTransaction{}}, nil
	}

	return &apiResponse, nil
}

//FetchPayout through External API Lazada
func (r *paymentsRepository) FetchPayouts(createdAfter string) (*models.LazadaAPIResponse, error) {
	// Initialize the map for query parameters
	queryParams := make(map[string]string)

	// Add parameters to the query map
	r.client.AddAPIParam("created_after", createdAfter)

	// API call
	resp, err := r.client.Execute("/finance/payout/status/get", "GET", queryParams)
	if err != nil {
		log.Println("Error fetching payout:", err)
		return nil, err
	}

	// Initialize the response struct
	var apiResponse models.LazadaAPIResponse
	err = json.Unmarshal(resp.Data, &apiResponse.Payout)
	if err != nil {
		log.Println("JSON Unmarshal Error:", err)
		return nil, err
	}

	if apiResponse.Payout == nil {
		log.Println("API response `data` field is missing or null")
		return &models.LazadaAPIResponse{Payout: []models.LazadaPayout{}}, nil
	}

	return &apiResponse, nil
}