
package repositories

import (
	"backend_project/internal/payment/models"
	"backend_project/sdk"
	"database/sql"
	"encoding/json"
	"errors"
)

type ReturnRepository interface {
	ProcessReturn(trade_order_id string, page_no string, page_size string) (models.ReturnData, error)
}

type returnRepository struct {
	client      *sdk.IopClient
	appKey      string
	accessToken string
	DB          *sql.DB
}

func NewReturnRepository(client *sdk.IopClient, appKey, accessToken string, db *sql.DB) ReturnRepository {
	return &returnRepository{client, appKey, accessToken, db}
}

func (r *returnRepository) ProcessReturn(trade_order_id string, page_no string, page_size string) (models.ReturnData, error) {
	queryParams := map[string]string{
		"appKey":      r.appKey,
		"accessToken": r.accessToken,
	}

	r.client.AddAPIParam("trade_order_id", trade_order_id)
	r.client.AddAPIParam("page_no", page_no)
	r.client.AddAPIParam("page_size", page_size)

	resp, err := r.client.Execute("/reverse/getreverseordersforseller", "GET", queryParams)
	if err != nil {
		return models.ReturnData{}, err
	}

	var returnData models.ReturnData
	err = json.Unmarshal(resp.Result, &returnData)
	if err != nil {
		return models.ReturnData{}, errors.New("failed to parse return data")
	}

	return returnData, nil
}
