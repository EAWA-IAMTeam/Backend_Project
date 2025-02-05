package models

import (
	"backend_project/internal/products/models"
)

type Order struct {
	ID           int            `json:"id"`
	Product      models.Product `json:"product"`
	Quantity     int            `json:"quantity"`
	CustomerName string         `json:"customer_name"`
}
