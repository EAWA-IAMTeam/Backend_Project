package services

import (
	"backend_project/internal/config"
	"backend_project/internal/products/models"
	"backend_project/internal/products/repositories"
	"backend_project/sdk"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/labstack/echo/v4"
)

type ProductService struct {
	ProductRepo *repositories.ProductRepository
}

func NewProductService(pr *repositories.ProductRepository) *ProductService {
	return &ProductService{ProductRepo: pr}
}

// FetchStockItemsByCompany retrieves stock items by company ID
func (ps *ProductService) FetchStockItemsByCompany(companyID int) ([]*models.StockItem, error) {
	return ps.ProductRepo.GetStockItemsByCompany(companyID)
}

// FetchProductsByStore retrieves products by store ID
func (ps *ProductService) FetchProductsByStore(storeID int) ([]*models.MergeProduct, error) {
	return ps.ProductRepo.GetProductsByStore(storeID)
}

// ParseProductRequest reads and parses the request body
func (ps *ProductService) ParseProductRequest(c echo.Context) (*models.ProductRequest, error) {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		return nil, err
	}

	c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

	var req models.ProductRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	if req.StoreID == 0 || len(req.Products) == 0 {
		return nil, errors.New("store_id and products are required")
	}

	return &req, nil
}

// InsertProducts inserts products into the database
func (ps *ProductService) InsertProducts(req *models.ProductRequest) (*models.InsertResult, error) {
	return ps.ProductRepo.InsertProductBatch(req.StoreID, req.Products)
}

// FetchFilteredProducts retrieves and processes filtered products
func (ps *ProductService) FetchFilteredProducts(storeID string) (map[string]interface{}, error) {
	// Fetch SKUs to remove from the database
	skusToRemove, err := ps.ProductRepo.GetStoreSkus(storeID)
	if err != nil {
		log.Printf("Failed to fetch SKUs: %v", err)
		return nil, fmt.Errorf("failed to retrieve SKU list")
	}

	// Initialize IOP client and Lazada API client
	lazadaClient, err := ps.initLazadaClient()
	if err != nil {
		return nil, err
	}

	// Fetch prolazadaClient external API
	resp, err := lazadaClient.Execute("/products/get", "GET", nil)
	if err != nil {
		log.Printf("Failed to fetch products from API: %v", err)
		return nil, fmt.Errorf("failed to fetch products")
	}

	// Process API response and filter products
	return ps.processApiResponse(resp, skusToRemove)
}

// initLazadaClient initializes and returns a Lazada API client
func (ps *ProductService) initLazadaClient() (*sdk.IopClient, error) {
	env := config.LoadConfig()
	clientOptions := sdk.ClientOptions{
		APIKey:    env.AppKey,
		APISecret: env.AppSecret,
		Region:    "MY", // Consider using a constant for region
	}
	lazadaClient := sdk.NewClient(&clientOptions)
	lazadaClient.SetAccessToken(env.AccessToken)
	return lazadaClient, nil
}

// processApiResponse filters products based on store SKUs
func (ps *ProductService) processApiResponse(resp interface{}, skusToRemove map[string]bool) (map[string]interface{}, error) {
	responseBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %v", err)
	}

	var apiResponse models.ApiResponse
	if err := json.Unmarshal(responseBytes, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	var unmappedProducts, mappedProducts []models.Product

	for _, product := range apiResponse.Data.Products {
		var remainingSkus, removedSkus []models.Sku

		for _, sku := range product.Skus {
			if skusToRemove[sku.ShopSku] {
				removedSkus = append(removedSkus, sku)
			} else {
				remainingSkus = append(remainingSkus, sku)
			}
		}

		if len(remainingSkus) > 0 {
			product.Skus = remainingSkus
			unmappedProducts = append(unmappedProducts, product)
		}
		if len(removedSkus) > 0 {
			productCopy := product
			productCopy.Skus = removedSkus
			mappedProducts = append(mappedProducts, productCopy)
		}
	}

	return map[string]interface{}{
		"unmapped_products": unmappedProducts,
		"mapped_products":   mappedProducts,
	}, nil
}