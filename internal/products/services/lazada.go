package services

import (
	"backend_project/database"
	"backend_project/internal/config"
	"backend_project/internal/products/models"
	"backend_project/sdk"
	"encoding/json"
	"fmt"
	"log"
)

// Get Access Token
func (ps *ProductService) GetAccessToken(storeID int64) string {
	db, err := database.ConnectDB()
	if err != nil {
		return ""
	}

	var accessToken string
	db.QueryRow("SELECT access_token FROM accesstoken WHERE store_id = $1", storeID).Scan(&accessToken)
	log.Println("Acess Token: ", accessToken)
	return accessToken

}

// initLazadaClient initializes and returns a Lazada API client
func (ps *ProductService) initLazadaClient(storeID int64) (*sdk.IopClient, error) {

	env := config.LoadConfig()
	clientOptions := sdk.ClientOptions{
		APIKey:    env.AppKey,
		APISecret: env.AppSecret,
		Region:    "MY", // Consider using a constant for region
	}
	lazadaClient := sdk.NewClient(&clientOptions)
	// lazadaClient.SetAccessToken(env.AccessToken)
	lazadaClient.SetAccessToken(ps.GetAccessToken(storeID))
	return lazadaClient, nil
}

// FetchMappedProducts retrieves products that are already mapped (removed from external API)
func (ps *ProductService) FetchLazadaMappedProducts(storeID int64) ([]models.Product, error) {
	skusToRemove, err := ps.ProductRepo.GetStoreSkus(storeID)
	if err != nil {
		log.Printf("Failed to fetch SKUs: %v", err)
		return nil, fmt.Errorf("failed to retrieve SKU list")
	}

	lazadaClient, err := ps.initLazadaClient(storeID)
	if err != nil {
		return nil, err
	}

	resp, err := lazadaClient.Execute("/products/get", "GET", nil)
	// log.Print(resp)
	if err != nil {
		log.Printf("Failed to fetch products from API: %v", err)
		return nil, fmt.Errorf("failed to fetch products")
	}

	return ps.filterLazadaMappedProducts(resp, storeID, skusToRemove)
}

// filterMappedProducts filters out products that are MAPPED (removed from external API)
func (ps *ProductService) filterLazadaMappedProducts(resp interface{}, storeID int64, skusToRemove map[string]bool) ([]models.Product, error) {
	responseBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %v", err)
	}

	var apiResponse models.ApiResponse
	if err := json.Unmarshal(responseBytes, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	var mappedProducts []models.Product

	for _, lazadaProduct := range apiResponse.Data.Products {
		var product models.Product
		product.ItemID = lazadaProduct.ItemID
		product.StoreID = storeID
		product.Name = lazadaProduct.Attributes.Name
		product.Description = lazadaProduct.Attributes.Description
		product.Images = lazadaProduct.Images

		var removedSkus []models.Sku
		for _, sku := range lazadaProduct.Skus {
			if skusToRemove[sku.ShopSku] {
				removedSkus = append(removedSkus, sku)
			}
		}

		if len(removedSkus) > 0 {
			productCopy := product
			productCopy.Skus = removedSkus
			mappedProducts = append(mappedProducts, productCopy)
		}
	}

	return mappedProducts, nil
}

// FetchUnmappedProducts retrieves products that are NOT mapped (still available in external API)
func (ps *ProductService) FetchLazadaUnmappedProducts(storeID int64) ([]models.Product, error) {
	skusToRemove, err := ps.ProductRepo.GetStoreSkus(storeID)
	if err != nil {
		log.Printf("Failed to fetch SKUs: %v", err)
		return nil, fmt.Errorf("failed to retrieve SKU list")
	}

	// TODO: Need to change the access token according to storeID to fetch the products of the store
	// TODO: Need to store the access token in the database
	lazadaClient, err := ps.initLazadaClient(storeID)
	if err != nil {
		return nil, err
	}

	resp, err := lazadaClient.Execute("/products/get", "GET", nil)
	if err != nil {
		log.Printf("Failed to fetch products from API: %v", err)
		return nil, fmt.Errorf("failed to fetch products")
	}

	return ps.filterLazadaUnmappedProducts(resp, storeID, skusToRemove)
}

// filterUnmappedProducts filters out products that are NOT mapped (available in external API)
func (ps *ProductService) filterLazadaUnmappedProducts(resp interface{}, storeID int64, skusToRemove map[string]bool) ([]models.Product, error) {
	responseBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %v", err)
	}

	var apiResponse models.ApiResponse
	if err := json.Unmarshal(responseBytes, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	var unmappedProducts []models.Product

	for _, lazadaProduct := range apiResponse.Data.Products {
		var product models.Product
		product.ItemID = lazadaProduct.ItemID
		product.StoreID = storeID
		product.Name = lazadaProduct.Attributes.Name
		product.Description = lazadaProduct.Attributes.Description
		product.Images = lazadaProduct.Images

		var remainingSkus []models.Sku

		for _, sku := range lazadaProduct.Skus {
			if !skusToRemove[sku.ShopSku] {
				remainingSkus = append(remainingSkus, sku)
			}
		}

		if len(remainingSkus) > 0 {
			productCopy := product
			productCopy.Skus = remainingSkus
			unmappedProducts = append(unmappedProducts, productCopy)
		}
	}

	return unmappedProducts, nil
}
