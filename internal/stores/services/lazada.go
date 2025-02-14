package services

import (
	"backend_project/internal/config"
	"backend_project/sdk"
	"fmt"

	"github.com/labstack/gommon/log"
)

// initLazadaClient initializes and returns a Lazada API client
func (ss *StoreService) initLazadaClient() (*sdk.IopClient, error) {
	env := config.LoadConfig()
	clientOptions := sdk.ClientOptions{
		APIKey:    env.AppKey,
		APISecret: env.AppSecret,
		Region:    "MY",
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

	resp, authResp, err := lazadaClient.Execute("/auth/token/create", "GET", nil)
	if err != nil {
		return "", fmt.Errorf("API request error: %v", err)
	}

	log.Printf("Lazada API response: %+v\n", resp)

	// Validate Lazada API response
	if resp.Code != "0" {
		return "", fmt.Errorf("lazada API Error: %s - %s", resp.Code, resp.Message)
	}

	if authResp == nil {
		return "", fmt.Errorf("failed to parse Lazada auth response")
	}

	return authResp.AccessToken, nil
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
