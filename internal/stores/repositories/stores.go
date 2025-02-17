package repositories

import (
	"backend_project/internal/stores/models"
	"database/sql"
	"fmt"
)

type StoreRepository interface {
	GetAccountByUserID(userID string) (*models.Account, error)
	SaveAccount(account *models.Account) error
	SaveStore(store *models.Store) error
	SaveAccessToken(accessToken *models.AccessToken) error
	UpdateStore(store *models.Store) error
}

type storeRepository struct {
	DB *sql.DB
}

func NewStoreRepository(db *sql.DB) StoreRepository {
	return &storeRepository{DB: db}
}

func (sr *storeRepository) GetAccountByUserID(userID string) (*models.Account, error) {
	query := `
	SELECT id, company_id, name, platform, region, is_main 
	FROM account 
	WHERE id = $1`

	row := sr.DB.QueryRow(query, userID)

	var account models.Account
	err := row.Scan(
		&account.ID,
		&account.CompanyID,
		&account.Name,
		&account.Platform,
		&account.Region,
		&account.IsMain)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No account found, return nil
		}
		return nil, fmt.Errorf("failed to get account by user ID: %v", err)
	}
	return &account, nil
}

func (sr *storeRepository) SaveAccount(account *models.Account) error {
	query := `
	INSERT INTO account (id, company_id, name, platform, region, is_main) 
	VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := sr.DB.Exec(query,
		account.ID,
		account.CompanyID,
		account.Name,
		account.Platform,
		account.Region,
		account.IsMain)
	if err != nil {
		return fmt.Errorf("failed to save account: %v", err)
	}
	return nil
}

func (sr *storeRepository) SaveStore(store *models.Store) error {
	query := `
	INSERT INTO store (id, company_id, access_token_id, name, platform, region, descriptions, status) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := sr.DB.Exec(query,
		store.ID,
		store.CompanyID,
		store.AccessTokenID,
		// store.ExpiryTime,
		store.Name,
		store.Platform,
		store.Region,
		store.Descriptions,
		store.Status)
	if err != nil {
		return fmt.Errorf("failed to save store: %v", err)
	}
	return nil
}

func (sr *storeRepository) SaveAccessToken(accessToken *models.AccessToken) error {
	query := `
	INSERT INTO accessToken (id, account_id, store_id, access_token, refresh_token, platform) 
	VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := sr.DB.Exec(query,
		accessToken.ID,
		accessToken.AccountID,
		accessToken.StoreID,
		accessToken.AccessToken,
		accessToken.RefreshToken,
		accessToken.Platform)
	if err != nil {
		return fmt.Errorf("failed to save access token: %v", err)
	}
	return nil
}

func (sr *storeRepository) UpdateStore(store *models.Store) error {
	query := `
	UPDATE store 
	SET access_token_id = $1 
	WHERE id = $2`
	_, err := sr.DB.Exec(query,
		store.AccessTokenID,
		store.ID)
	if err != nil {
		return fmt.Errorf("failed to update store: %v", err)
	}
	return nil
}
