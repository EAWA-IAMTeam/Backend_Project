package repositories

import (
	"backend_project/internal/stores/models"
	"database/sql"
	"fmt"
)

type StoreRepository interface {
	SaveStoreInfo(storeInfo models.Store) error
	GetStoreByID(storeID int64) (*models.Store, error)
}

type storeRepository struct {
	DB *sql.DB
}

func NewStoreRepository(db *sql.DB) StoreRepository {
	return &storeRepository{DB: db}
}

func (sr *storeRepository) SaveStoreInfo(storeInfo models.Store) error {
	// Implement the logic to save store info to the database
	query := "INSERT INTO stores (id, company_id, access_token_id, expiry_time, name, platform, region, description, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := sr.DB.Exec(query, storeInfo.ID, storeInfo.CompanyID, storeInfo.AccessTokenID, storeInfo.ExpiryTime, storeInfo.Name, storeInfo.Platform, storeInfo.Region, storeInfo.Description, storeInfo.Status)
	if err != nil {
		return fmt.Errorf("failed to save store info: %v", err)
	}
	return nil
}

func (sr *storeRepository) GetStoreByID(storeID int64) (*models.Store, error) {
	// Implement the logic to get store info by ID from the database
	query := "SELECT id, company_id, access_token_id, expiry_time, name, platform, region, description, status FROM stores WHERE id = ?"
	row := sr.DB.QueryRow(query, storeID)

	var store models.Store
	err := row.Scan(&store.ID, &store.CompanyID, &store.AccessTokenID, &store.ExpiryTime, &store.Name, &store.Platform, &store.Region, &store.Description, &store.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get store info: %v", err)
	}
	return &store, nil
}
