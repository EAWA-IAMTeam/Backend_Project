package services

import (
	"backend_project/internal/products/models"
	"backend_project/internal/products/repositories"
	"bytes"
	"encoding/json"
	"errors"
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
func (ps *ProductService) FetchProductsByCompany(companyID int) ([]*models.MergeProduct, error) {
	return ps.ProductRepo.GetProductsByCompany(companyID)
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
