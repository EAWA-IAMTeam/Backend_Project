package services

import (
	"backend_project/internal/config"
	"backend_project/sdk"
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

func (ss *StoreService) LazadaGenerateAccessToken(authCode string)
