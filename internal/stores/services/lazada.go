package services

import (
	"backend_project/internal/config"
	"backend_project/internal/stores/models"
	"backend_project/sdk"
	"fmt"

	"github.com/labstack/gommon/log"
)

// initLazadaClient initializes and returns a Lazada API client
func (ss storeService) initLazadaClient() (*sdk.IopClient, error) {
	env := config.LoadConfig()
	clientOptions := sdk.ClientOptions{
		APIKey:    env.AppKey,
		APISecret: env.AppSecret,
		Region:    "MY",
	}
	lazadaClient := sdk.NewClient(&clientOptions)
	return lazadaClient, nil
}

// Call Lazada API to generate access token, and return the response in LinkStore format
func (ss *storeService) LazadaGenerateAccessToken(authCode string) (*models.LinkStore, error) {
	lazadaClient, err := ss.initLazadaClient()
	if err != nil {
		return nil, err
	}

	lazadaClient.AddAPIParam("code", authCode)

	resp, authResp, err := lazadaClient.Execute("/auth/token/create", "GET", nil)
	if err != nil {
		return nil, fmt.Errorf("API request error: %v", err)
	}

	log.Printf("Lazada API response: %+v\n", resp)

	// Validate Lazada API response
	if resp.Code != "0" {
		return nil, fmt.Errorf("lazada API Error: %s - %s", resp.Code, resp.Message)
	}

	// Extract the first UserInfo entry (assuming there's at least one)
	var userInfo models.UserInfo
	if len(authResp.UserInfo) > 0 {
		userInfo = authResp.UserInfo[0]
	}

	// Map the response to LinkStore
	linkStore := &models.LinkStore{
		AccessToken:      authResp.AccessToken,
		ExpiresIn:        authResp.ExpiresIn,
		RefreshToken:     authResp.RefreshToken,
		RefreshExpiresIn: authResp.RefreshExpiresIn,
		Country:          authResp.Country,
		UserID:           userInfo.UserID,
		SellerID:         userInfo.SellerID,
		Account:          authResp.Account,
		ShortCode:        userInfo.ShortCode,
	}

	return linkStore, nil
}

// TODO: Modify the return into models.ApiResponseStoreInfo
func (ss storeService) LazadaFetchStoreInfo(accessToken string) (interface{}, error) {

	lazadaClient, err := ss.initLazadaClient()
	if err != nil {
		return nil, err
	}

	lazadaClient.AddAPIParam("access_token", accessToken)

	storeInfo, _, err := lazadaClient.Execute("/seller/get", "GET", nil)
	if err != nil {
		return nil, err
	}

	return storeInfo, nil
}
