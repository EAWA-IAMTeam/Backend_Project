package services

import (
	"backend_project/internal/products/models"
	"backend_project/internal/products/repositories"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"strconv"

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
func (ps *ProductService) FetchProductsByCompany(companyID int) ([]*models.MergeProduct, error) {
	return ps.ProductRepo.GetProductsByCompany(companyID)
}

// ParseProductRequest reads and parses the request body
func (ps *ProductService) ParseProductRequest(c echo.Context) ([]*models.StoreProduct, error) {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		return nil, err
	}

	c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

	var req []*models.StoreProduct
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	return req, nil
}

// InsertProducts inserts products into the database
func (ps *ProductService) InsertProducts(req []*models.StoreProduct) (*models.InsertResult, error) {
	return ps.ProductRepo.InsertProductBatch(req)
}

// FetchMappedProducts retrieves products that are already mapped (removed from external API)
func (ps *ProductService) FetchMappedProducts(companyID string) ([]models.Product, error) {
	// TODO: get mapped products from all platforms and return
	var products []models.Product

	storeIDs, err := ps.ProductRepo.GetStoreByCompany(companyID)
	if err != nil {
		return nil, err
	}
	
	for _, storeID := range storeIDs["Lazada"] {
		lazadaProducts, err := ps.FetchLazadaMappedProducts(strconv.FormatInt(storeID, 10))
		if err != nil {
			return nil, err
		}
		products = append(products, lazadaProducts...) // Spread the slice correctly
	}


	// TODO: Implement another 2 platform as following by refering to Lazada

	// TODO: Implement FetchShopeeMappedProducts
	// for _, storeID := range storeIDs["Shopee"] {
	// 	shopeeProducts, err := ps.FetchShopeeMappedProducts(strconv.FormatInt(storeID, 10))
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	products = append(products, shopeeProducts...) // Spread the slice correctly
	// }

	// TODO: Implement FetchTikTokMappedProducts
	// for _, storeID := range storeIDs["TikTok"] {
	// 	tiktokProducts, err := ps.FetchTikTokMappedProducts(strconv.FormatInt(storeID, 10))
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	products = append(products, tiktokProducts...) // Spread the slice correctly
	// }
	
	if err != nil {
		return nil, err
	}
	return products, nil
}

// FetchMappedProducts retrieves products that are already mapped (removed from external API)
func (ps *ProductService) FetchUnmappedProducts(companyID string) ([]models.Product, error) {
	// TODO: get mapped products from all platforms and return
	var products []models.Product

	storeIDs, err := ps.ProductRepo.GetStoreByCompany(companyID)
	if err != nil {
		return nil, err
	}
	
	for _, storeID := range storeIDs["Lazada"] {
		lazadaProducts, err := ps.FetchLazadaUnmappedProducts(strconv.FormatInt(storeID, 10))
		if err != nil {
			return nil, err
		}
		products = append(products, lazadaProducts...) // Spread the slice correctly
	}


	// TODO: Implement another 2 platform as following by refering to Lazada

	// TODO: Implement FetchShopeeMappedProducts
	// for _, storeID := range storeIDs["Shopee"] {
	// 	shopeeProducts, err := ps.FetchShopeeUnmappedProducts(strconv.FormatInt(storeID, 10))
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	products = append(products, shopeeProducts...) // Spread the slice correctly
	// }

	// TODO: Implement FetchTikTokMappedProducts
	// for _, storeID := range storeIDs["TikTok"] {
	// 	tiktokProducts, err := ps.FetchTikTokUnmappedProducts(strconv.FormatInt(storeID, 10))
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	products = append(products, tiktokProducts...) // Spread the slice correctly
	// }
	
	if err != nil {
		return nil, err
	}
	return products, nil
}

// Delete products that are MAPPED
func (ps *ProductService) DeleteMappedProducts(storeID string, sku string) (int64, error) {
	return ps.ProductRepo.DeleteMappedProductsBySKU(storeID, sku)
}

// DeleteMappedProductsBatch deletes multiple mapped products and returns successfully deleted and failed SKUs
func (ps *ProductService) DeleteMappedProductsBatch(storeID string, skus []string) ([]string, []string, error) {
	deletedSKUs, failedSKUs, err := ps.ProductRepo.DeleteMappedProductsBySKUs(storeID, skus)
	if err != nil {
		return nil, nil, err
	}

	return deletedSKUs, failedSKUs, nil
}
