package services

import (
	"backend_project/internal/config"
	"backend_project/sdk"
	"fmt"
)

// initLazadaClient initializes and returns a Lazada API client
func (ss *StoreService) initLazadaClient() (*sdk.IopClient, error) {
	env := config.LoadConfig()
	clientOptions := sdk.ClientOptions{
		APIKey:    env.AppKey,
		APISecret: env.AppSecret,
		Region:    "MY", // Consider using a constant for region
	}
	lazadaClient := sdk.NewClient(&clientOptions)
	return lazadaClient, nil
}

func (ss *StoreService) LazadaGenerateAccessToken(authCode string) (string, error) {
	lazadaClient, err := ss.initLazadaClient()
	if err != nil {
		return "", err
	}

	lazadaClient.AddAPIParam("code", authCode)

	resp, _, err := lazadaClient.Execute("/auth/token/create", "GET", nil)
	if err != nil {
		return "", fmt.Errorf("API request error: %v", err)
	}

	// Validate Lazada API response
	if resp.Code != "0" {
		return "", fmt.Errorf("Lazada API Error: %s - %s", resp.Code, resp.Message)
	}

	return resp.Message, nil
}

func (ss *StoreService) LazadaFetchStoreInfo(accessToken string) (interface{}, error) {

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
