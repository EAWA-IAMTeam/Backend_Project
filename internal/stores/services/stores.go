package services

import (
	"backend_project/internal/stores/models"
	"backend_project/internal/stores/repositories"
	"fmt"
	"log"
)

/*
TODO:
1. Call Lazada 'auth/token/create' endpoint to generate access token. [done]
2. Extract the access token from the response. [done]
3. Call Lazada 'seller/get' endpoint to fetch store info using the access token. [done]
4. check user_id and seller_id from the response are same or not [for isMain purpose under account table] [done]
5. check the user_id is already exist in the database or not [done]
6. if not exist, create a new record in the account table. [done]
7.  if the store was exist, then update the name and access token or what.


Exception ways:
for step 1: might get response with error code. [need to validate the response first before return the response]
same goes to step 3. [done]

Importance:
1. set prefix e-commerce platform name for the id for store and account table. example: Lazada239827329 [done]
2. need to create a new record in the account and store table, only can insert access token info into accessToken table

Problem Facing:
1. how to receive company id? pass on body?
2. The store expiry time issues [need to be in timestampz]
3. accessToken ID set in autoincrement or backend set ourself.
*/

type StoreService interface {
	FetchStoreInfo(authCode string) (interface{}, error)
}

type storeService struct {
	storeRepository repositories.StoreRepository
}

func NewStoreService(sr repositories.StoreRepository) StoreService {
	return &storeService{storeRepository: sr}
}

func (ss *storeService) FetchStoreInfo(authCode string) (interface{}, error) {
	linkStore, err := ss.LazadaGenerateAccessToken(authCode)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %v", err)
	}

	storeInfo, err := ss.LazadaFetchStoreInfo(linkStore.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch store info: %v", err)
	}

	isMain := linkStore.UserID == linkStore.SellerID

	existingAccount, err := ss.storeRepository.GetAccountByUserID("Lazada" + linkStore.UserID)
	if err != nil {
		log.Printf("Warning: failed to check existing account: %v. Proceeding to create a new account.", err)
		existingAccount = nil
	}

	var accountID string
	if existingAccount == nil {
		account := &models.Account{
			ID:        "Lazada" + linkStore.UserID,
			CompanyID: 0,
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

	store := &models.Store{
		ID:            "Lazada" + linkStore.SellerID,
		CompanyID:     0,
		AccessTokenID: 0,
		// ExpiryTime:    "",
		Name:         storeInfo.Name,
		Platform:     "Lazada",
		Region:       linkStore.Country,
		Descriptions: "",
		Status:       true,
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
