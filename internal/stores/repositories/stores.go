package repositories

import (
	"backend_project/internal/stores/models"
	"database/sql"
	"fmt"
	"time"
)

type StoreRepository interface {
	GetAccountByUserID(userID string) (*models.Account, error)
	SaveAccount(account *models.Account) error
	SaveStore(store *models.Store) error
	SaveAccessToken(accessToken *models.AccessToken) error
	UpdateStore(store *models.Store) error
	GetStoresByCompany(companyID int64) ([]*models.Store, error)
}

type storeRepository struct {
	DB *sql.DB
}

func NewStoreRepository(db *sql.DB) StoreRepository {
	return &storeRepository{DB: db}
}

// check if an account already exists in the database
// If the account exists, it is returned.
// If the account does not exist (nil is returned), the next step is to create a new account.
func (sr *storeRepository) GetAccountByUserID(userID string) (*models.Account, error) {
	fmt.Printf("userID: %v\n", userID)
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

// If the account does not exist, a new models.Account struct is created with the necessary details.
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

// func (sr *storeRepository) SaveStore(store *models.Store) error {
// 	query := `
// 	INSERT INTO store (id, company_id, access_token_id, authorize_time, expiry_time, name, platform, region, descriptions, status)
// 	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
// 	_, err := sr.DB.Exec(query,
// 		store.ID,
// 		store.CompanyID,
// 		store.AccessTokenID,
// 		store.AuthTime,
// 		store.ExpiryTime,
// 		store.Name,
// 		store.Platform,
// 		store.Region,
// 		store.Descriptions,
// 		store.Status)
// 	if err != nil {
// 		return fmt.Errorf("failed to save store: %v", err)
// 	}
// 	return nil
// }

func (sr *storeRepository) SaveStore(store *models.Store) error {
	// Check if store ID exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM store WHERE id = $1)`
	err := sr.DB.QueryRow(checkQuery, store.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if store exists: %v", err)
	}

	if exists {
		// If store exists, update authorize_time, expiry_time, and status
		updateQuery := `
		UPDATE store 
		SET authorize_time = $1, expiry_time = $2, status = $3 
		WHERE id = $4`
		_, err = sr.DB.Exec(updateQuery, store.AuthTime, store.ExpiryTime, store.Status, store.ID)
		if err != nil {
			return fmt.Errorf("failed to update store: %v", err)
		}
	} else {
		// If store does not exist, insert a new record
		insertQuery := `
		INSERT INTO store (id, company_id, access_token_id, authorize_time, expiry_time, name, platform, region, descriptions, status) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
		_, err = sr.DB.Exec(insertQuery,
			store.ID,
			store.CompanyID,
			store.AccessTokenID,
			store.AuthTime,
			store.ExpiryTime,
			store.Name,
			store.Platform,
			store.Region,
			store.Descriptions,
			store.Status)
		if err != nil {
			return fmt.Errorf("failed to insert store: %v", err)
		}
	}

	return nil
}

// if store id exist then update access token and refresh token only, not insert
func (sr *storeRepository) SaveAccessToken(accessToken *models.AccessToken) error {
	// Directly check if the record exists and get the ID
	checkQuery := `SELECT id FROM accessToken WHERE store_id = $1`
	err := sr.DB.QueryRow(checkQuery, accessToken.StoreID).Scan(&accessToken.ID)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing access token: %v", err)
	}

	if accessToken.ID > 0 {
		// If exists, update and return the same ID
		updateQuery := `
		UPDATE accessToken 
		SET access_token = $1, refresh_token = $2 
		WHERE id = $3
		RETURNING id`
		err = sr.DB.QueryRow(updateQuery, accessToken.AccessToken, accessToken.RefreshToken, accessToken.ID).Scan(&accessToken.ID)
		if err != nil {
			return fmt.Errorf("failed to update access token: %v", err)
		}
	} else {
		// If not exists, insert new and get the ID
		insertQuery := `
		INSERT INTO accessToken (account_id, store_id, access_token, refresh_token, platform) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id`
		err = sr.DB.QueryRow(insertQuery,
			accessToken.AccountID,
			accessToken.StoreID,
			accessToken.AccessToken,
			accessToken.RefreshToken,
			accessToken.Platform,
		).Scan(&accessToken.ID)

		if err != nil {
			return fmt.Errorf("failed to insert access token: %v", err)
		}
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

// GetOrdersByCompany fetches orders by company ID
func (sr *storeRepository) GetStoresByCompany(companyID int64) ([]*models.Store, error) {

	// Get paginated data
	query := `
	SELECT id, company_id, access_token_id, authorize_time, expiry_time, 
			name, platform, region, descriptions, status
	FROM store 
	WHERE company_id = $1
	ORDER BY authorize_time DESC`

	rows, err := sr.DB.Query(query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stores: %v", err)
	}
	defer rows.Close()

	var stores []*models.Store

	now := time.Now() // Get current time

	for rows.Next() {
		var store models.Store
		err := rows.Scan(
			&store.ID,
			&store.CompanyID,
			&store.AccessTokenID,
			&store.AuthTime,
			&store.ExpiryTime,
			&store.Name,
			&store.Platform,
			&store.Region,
			&store.Descriptions,
			&store.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan store: %v", err)
		}

		// Check if the access token is expired
		if store.ExpiryTime.Before(now) {
			store.Status = false // Set to inactive

			// Update the database to reflect the status change
			updateQuery := `UPDATE store SET status = false WHERE id = $1`
			_, updateErr := sr.DB.Exec(updateQuery, store.ID)
			if updateErr != nil {
				return nil, fmt.Errorf("failed to update store status: %v", updateErr)
			}
		}

		stores = append(stores, &store)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through stores: %v", err)
	}

	return stores, nil
}
