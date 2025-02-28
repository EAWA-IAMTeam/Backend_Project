package services

import (
	"backend_project/internal/stores/models"
	"backend_project/internal/stores/repositories"
	"fmt"
	"time"
	//"log"
)

type StoreService interface {
	FetchStoreInfo(authCode string, companyID int64) (interface{}, error)
	GetStoresByCompany(companyID int64) ([]*models.Store, error)
}

type storeService struct {
	storeRepository repositories.StoreRepository
}

// create instances
func NewStoreService(sr repositories.StoreRepository) StoreService {
	return &storeService{storeRepository: sr}
}

// Implement GetStoresByCompany
func (ss *storeService) GetStoresByCompany(companyID int64) ([]*models.Store, error) {
	//return ss.storeRepository.GetStoresByCompany(companyID) // Call the repository method
	stores, err := ss.storeRepository.GetStoresByCompany(companyID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for _, store := range stores {
		// Only update status or tokens for active stores
		if store.Status {

			// Check login time if the current date is 30 days from the AuthTime
			if now.Sub(store.AuthTime) >= 30*24*time.Hour {
				store.Status = false
				err = ss.storeRepository.UpdateStoreStatus(store)
				if err != nil {
					return nil, fmt.Errorf("failed to update store status: %v", err)
				}
				continue // Skip further checks for this store
			}

			if store.ExpiryTime.Before(now) { // Access token expired
				if store.RefreshExpiryTime.After(now) { // Refresh token still valid
					//get refresh token from access token table, pass back access token
					token, err1 := ss.storeRepository.GetTokenByStoreID(store.StoreID)
					if err1 != nil {
						return nil, fmt.Errorf("failed to get token: %v", err)
					}
					// Call API to refresh access , pass back linkstore
					newToken, err := ss.LazadaRefreshToken(token.RefreshToken)
					if err != nil {
						return nil, fmt.Errorf("failed to refresh access token: %v", err)
					}
					token.AccessToken = newToken.AccessToken
					token.RefreshToken = newToken.RefreshToken

					//update access token and refresh token
					err = ss.storeRepository.SaveAccessToken(token)

					if err != nil {
						return nil, fmt.Errorf("failed to update store tokens: %v", err)
					}
					// Update tokens in DB
					store.ExpiryTime = now.Add(time.Second * time.Duration(604800)) // 7 days later
					store.RefreshExpiryTime = now.Add(time.Second * time.Duration(2592000))
					err = ss.storeRepository.SaveStore(store)

					if err != nil {
						return nil, fmt.Errorf("failed to update store info: %v", err)
					}
				}
			}
		}
	}

	return stores, nil
}

// If FetchStoreInfo sometimes returns different data structures, using interface{} avoids type restrictions.
func (ss *storeService) FetchStoreInfo(authCode string, companyID int64) (interface{}, error) {
	linkStore, err := ss.LazadaGenerateAccessToken(authCode)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %v", err)
	}

	storeInfo, err := ss.LazadaFetchStoreInfo(linkStore.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch store info: %v", err)
	}

	isMain := linkStore.UserID == linkStore.SellerID

	//existingAccount, err := ss.storeRepository.GetAccountByUserID("Lazada" + linkStore.UserID)
	existingAccount, err := ss.storeRepository.GetAccountByUserID(linkStore.UserID)
	if err != nil {
		//log.Printf("Warning: failed to check existing account: %v. Proceeding to create a new account.", err)
		existingAccount = nil
	}

	var accountID string
	if existingAccount == nil {
		account := &models.Account{
			//ID:        "Lazada" + linkStore.UserID,
			ID:        0, // Auto-incremented ID (if applicable)
			AccountID: linkStore.UserID,
			CompanyID: companyID,
			Name:      linkStore.Account,
			Platform:  "Lazada",
			Region:    linkStore.Country,
			IsMain:    isMain,
		}

		err = ss.storeRepository.SaveAccount(account)
		if err != nil {
			return nil, fmt.Errorf("failed to save account: %v", err)
		}

		accountID = account.AccountID
	} else {
		accountID = existingAccount.AccountID
	}
	//authorization time
	authTime := time.Now()

	expiresAt := authTime.Add(time.Second * time.Duration(604800))         // 7 days later
	refreshExpiresAt := authTime.Add(time.Second * time.Duration(2592000)) // 30 days later
	// expiresAt := authTime.Add(time.Minute * 1)
	// refreshExpiresAt := authTime.Add(time.Minute * 5)

	store := &models.Store{
		//ID:            "Lazada" + linkStore.SellerID,
		ID:                0, // Auto-incremented ID (if applicable)
		StoreID:           linkStore.SellerID,
		CompanyID:         companyID,
		AccessTokenID:     0,
		AuthTime:          authTime,
		ExpiryTime:        expiresAt,
		RefreshExpiryTime: refreshExpiresAt,
		Name:              storeInfo.Name,
		Platform:          "Lazada",
		Region:            linkStore.Country,
		Descriptions:      "",
		Status:            true,
	}

	err = ss.storeRepository.SaveStore(store)
	if err != nil {
		return nil, fmt.Errorf("failed to save store: %v", err)
	}

	accessToken := &models.AccessToken{
		ID:           0, // Auto-incremented ID (if applicable)
		AccountID:    accountID,
		StoreID:      store.StoreID,
		AccessToken:  linkStore.AccessToken,
		RefreshToken: linkStore.RefreshToken,
		Platform:     "Lazada",
	}

	err = ss.storeRepository.SaveAccessToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to save access token: %v", err)
	}

	store.AccessTokenID = accessToken.ID
	err = ss.storeRepository.UpdateStore(store)
	if err != nil {
		return nil, fmt.Errorf("failed to update store with access token ID: %v", err)
	}

	return storeInfo, nil
}
