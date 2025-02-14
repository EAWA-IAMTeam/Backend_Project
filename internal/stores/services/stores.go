package services

import (
	"backend_project/internal/stores/repositories"
	"backend_project/sdk"
	"fmt"
)

// type StoreService interface {
// 	FetchStoreInfo(authCode string) (interface{}, error)
// }

type StoreService struct {
	// *Service
	storerepository *repositories.StoreRepository
	iopClient       *sdk.IopClient
}

func NewStoreService(sr *repositories.StoreRepository, client *sdk.IopClient) *StoreService {
	return &StoreService{sr, client}
}

func (ss *StoreService) FetchStoreInfo(authCode string) (interface{}, error) {
	// Step 1: Exchange auth code for an access token
	ss.iopClient.AddAPIParam("code", authCode)

	tokenResponse, _, err := ss.iopClient.Execute("/auth/token/create", "POST", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain access token: %v", err)
	}

	// Extract token from response
	accessToken, ok := tokenResponse["access_token"].(string)
	if !ok || accessToken == "" {
		return nil, fmt.Errorf("invalid access token response")
	}

	// Step 2: Fetch Store Information using the obtained token
	ss.iopClient.SetAccessToken(accessToken)
	storeInfo, _, err := ss.iopClient.Execute("/seller/get", "GET", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch store info: %v", err)
	}

	// Step 3: Save store info to the database
	err = ss.storerepository.SaveStoreInfo(storeInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to save store info: %v", err)
	}

	return storeInfo, nil
}
