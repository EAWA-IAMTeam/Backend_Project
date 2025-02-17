package services

import (
	"backend_project/internal/stores/repositories"
	"fmt"
)

/*
TODO:
1. Call Lazada 'auth/token/create' endpoint to generate access token. [done]
2. Extract the access token from the response. [done]
3. Call Lazada 'seller/get' endpoint to fetch store info using the access token. [done]
4. check user_id and seller_id from the response are same or not [for isMain purpose under account table]
5. check the user_id is already exist in the database or not
6. if not exist, create a new record in the account table.


Exception ways:
for step 1: might get response with error code. [need to validate the response first before return the response]
same goes to step 3. [done]

Importance:
1. set prefix e-commerce platform name for the id for store and account table. example: Lazada239827329
2. need to create a new record in the account and store table, only can insert access token info into accessToken table

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

	resp, err := ss.LazadaGenerateAccessToken(authCode)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %v", err)
	}

	storeInfo, err := ss.LazadaFetchStoreInfo(resp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch store info: %v", err)
	}

	fmt.Println(storeInfo)

	// err = ss.storeRepository.SaveStoreInfo(storeInfo)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to save store info: %v", err)
	// }

	return storeInfo, nil
}
