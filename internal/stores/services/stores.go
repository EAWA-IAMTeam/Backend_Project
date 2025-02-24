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
	return ss.storeRepository.GetStoresByCompany(companyID) // Call the repository method
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
			ID:        linkStore.UserID,
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

		accountID = account.ID
	} else {
		accountID = existingAccount.ID
	}
	//authprization time
	authTime := time.Now()
	//expiration time
	expiresAt := authTime.Add(time.Second * time.Duration(604800)) // 7 days later

	store := &models.Store{
		//ID:            "Lazada" + linkStore.SellerID,
		ID:            linkStore.SellerID,
		CompanyID:     companyID,
		AccessTokenID: 0,
		AuthTime:      authTime,
		ExpiryTime:    expiresAt,
		Name:          storeInfo.Name,
		Platform:      "Lazada",
		Region:        linkStore.Country,
		Descriptions:  "",
		Status:        true,
	}

	err = ss.storeRepository.SaveStore(store)
	if err != nil {
		return nil, fmt.Errorf("failed to save store: %v", err)
	}

	accessToken := &models.AccessToken{
		ID:           0, // Auto-incremented ID (if applicable)
		AccountID:    accountID,
		StoreID:      store.ID,
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
