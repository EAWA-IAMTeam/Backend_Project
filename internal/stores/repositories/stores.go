package repositories

import "database/sql"

type StoreRepository struct {
	DB *sql.DB
}

func NewStoreRepository(db *sql.DB) *StoreRepository {
	return &StoreRepository{DB: db}
}

func (sr *StoreRepository) SaveStoreInfo(storeInfo interface{}) error {
	// Save store info to the database
	return nil
}
