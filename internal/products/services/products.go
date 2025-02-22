package services

import (
	"backend_project/internal/products/models"
	"backend_project/internal/products/repositories"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

type ProductService struct {
	ProductRepo *repositories.ProductRepository
}

func NewProductService(pr *repositories.ProductRepository) *ProductService {
	return &ProductService{ProductRepo: pr}
}

// FetchStockItemsByCompany retrieves stock items by company ID
func (ps *ProductService) FetchStockItemsByCompany(companyID int64, page, limit int) ([]*models.StockItem, error) {
	return ps.ProductRepo.GetStockItemsByCompany(companyID, page, limit)
}

// CreateStockItemsByCompany inserts stock items for a specific company
func (ps *ProductService) CreateStockItemsByCompany(companyID int64, stockItems []models.StockItem) error {
	return ps.ProductRepo.InsertStockItemsByCompany(companyID, stockItems)
}

// FetchProductsByStore retrieves products by store ID
func (ps *ProductService) FetchProductsByCompany(companyID int64, page, limit int) ([]*models.MergeProduct, error) {
	return ps.ProductRepo.GetProductsByCompany(companyID, page, limit)
}

// ParseProductRequest reads and parses the request body
func (ps *ProductService) ParseProductRequest(msg *nats.Msg) ([]*models.Request, error) {
	body := msg.Data

	// Log the received data for debugging (optional)
	log.Printf("Received NATS message: %s", string(body))

	// Parse the JSON payload
	var req []*models.Request
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Failed to parse JSON: %v", err)
		return nil, err
	}

	return req, nil
}

// InsertProducts inserts products into the database
func (ps *ProductService) InsertProducts(req []*models.Request) (*models.InsertResult, error) {
	return ps.ProductRepo.InsertProductBatch(req)
}

// FetchMappedProducts retrieves products that are already mapped (removed from external API)
// TODO: get mapped products from all platforms and return
func (ps *ProductService) FetchMappedProducts(companyID int64) ([]models.Product, error) {
	var products []models.Product

	storeIDs, err := ps.ProductRepo.GetStoreByCompany(companyID)
	if err != nil {
		return nil, err
	}

	for _, storeID := range storeIDs["Lazada"] {
		lazadaProducts, err := ps.FetchLazadaMappedProducts(storeID)
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

	// if err != nil {
	// 	return nil, err
	// }
	return products, nil
}

// FetchMappedProducts retrieves products that are already mapped (removed from external API)
// TODO: get mapped products from all platforms and return
func (ps *ProductService) FetchUnmappedProducts(companyID int64) ([]models.Product, error) {
	var products []models.Product

	storeIDs, err := ps.ProductRepo.GetStoreByCompany(companyID)
	if err != nil {
		log.Print("a")
		return nil, err
	}
	log.Print("Store:", storeIDs)

	// TODO: Fetch the products from all platforms according to the company's store by using the access token in database
	for _, storeID := range storeIDs["Lazada"] {
		lazadaProducts, err := ps.FetchLazadaUnmappedProducts(storeID)
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

	// if err != nil {
	// 	return nil, err
	// }
	return products, nil
}

// Delete products that are MAPPED
func (ps *ProductService) DeleteMappedProducts(storeID int64, sku string) (int64, error) {
	return ps.ProductRepo.DeleteMappedProductsBySKU(storeID, sku)
}

// DeleteMappedProductsBatch deletes multiple mapped products and returns successfully deleted and failed SKUs
func (ps *ProductService) DeleteMappedProductsBatch(storeID int64, skus []string) ([]string, []string, error) {
	deletedSKUs, failedSKUs, err := ps.ProductRepo.DeleteMappedProductsBySKUs(storeID, skus)
	if err != nil {
		return nil, nil, err
	}

	return deletedSKUs, failedSKUs, nil
}
