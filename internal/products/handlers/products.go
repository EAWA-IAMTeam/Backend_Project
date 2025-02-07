package handlers

import (
	"backend_project/internal/products/services"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type ProductHandler struct {
	ProductService *services.ProductService
}

func NewProductHandler(ps *services.ProductService) *ProductHandler {
	return &ProductHandler{ProductService: ps}
}

// GetStockItemsByCompany handles fetching stock items by company ID
func (ph *ProductHandler) GetStockItemsByCompany(c echo.Context) error {
	companyID, err := strconv.Atoi(c.Param("company_id"))
	if err != nil {
		log.Printf("Failed to convert company_id to int: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid company_id"})
	}

	stockItems, err := ph.ProductService.FetchStockItemsByCompany(companyID)
	if err != nil {
		log.Printf("Failed to fetch stock items: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to fetch stock items"})
	}

	return c.JSON(http.StatusOK, stockItems)
}

// GetProductsByStore handles fetching products by store ID
func (ph *ProductHandler) GetProductsByCompany(c echo.Context) error {
	companyID, err := strconv.Atoi(c.Param("company_id"))
	if err != nil {
		log.Printf("Failed to convert company_id to int: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid company_id"})
	}

	products, err := ph.ProductService.FetchProductsByCompany(companyID)
	if err != nil {
		log.Printf("Failed to fetch products: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to fetch products"})
	}

	return c.JSON(http.StatusOK, products)
}

// InsertProducts handles inserting products into the database
func (ph *ProductHandler) InsertProducts(c echo.Context) error {
	req, err := ph.ProductService.ParseProductRequest(c)
	if err != nil {
		log.Printf("Invalid request format: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request format"})
	}

	result, err := ph.ProductService.InsertProducts(req)
	if err != nil {
		log.Printf("Error inserting products: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to insert products"})
	}

	statusCode := http.StatusCreated
	if len(result.Duplicates) > 0 {
		statusCode = http.StatusConflict
	}

	return c.JSON(statusCode, result)
}

// GetMappedProducts handles API requests for mapped products
func (ph *ProductHandler) GetMappedProducts(c echo.Context) error {
	companyID := c.Param("company_id")
	if companyID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "company_id is required"})
	}

	products, err := ph.ProductService.FetchMappedProducts(companyID)
	if err != nil {
		log.Printf("Error fetching mapped products: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to fetch mapped products"})
	}

	return c.JSON(http.StatusOK, products)
}

// GetUnmappedProducts handles API requests for mapped products
func (ph *ProductHandler) GetUnmappedProducts(c echo.Context) error {
	companyID := c.Param("company_id")
	if companyID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "company_id is required"})
	}

	products, err := ph.ProductService.FetchUnmappedProducts(companyID)
	if err != nil {
		log.Printf("Error fetching mapped products: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to fetch mapped products"})
	}

	return c.JSON(http.StatusOK, products)
}

// RemoveMappedProducts handles API requests to delete mapped products
func (ph *ProductHandler) RemoveMappedProducts(c echo.Context) error {
	storeID := c.Param("store_id")
	sku := c.Param("sku")

	if storeID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "store_id is required"})
	}

	if sku == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "sku is required"})
	}

	rowsAffected, err := ph.ProductService.DeleteMappedProducts(storeID, sku)
	if err != nil {
		log.Printf("Error removing mapped product (store_id: %s, sku: %s): %v", storeID, sku, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to remove mapped product"})
	}

	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Product not found or already removed"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Mapped product successfully removed"})
}

// RemoveMappedProductsBatch handles API requests to delete multiple mapped products
func (ph *ProductHandler) RemoveMappedProductsBatch(c echo.Context) error {
	storeID := c.Param("store_id")

	if storeID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "store_id is required"})
	}

	var request struct {
		SKUs []string `json:"skus"`
	}

	// Parse JSON body
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
	}

	if len(request.SKUs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "At least one SKU is required"})
	}

	deletedSKUs, failedSKUs, err := ph.ProductService.DeleteMappedProductsBatch(storeID, request.SKUs)
	if err != nil {
		log.Printf("Error removing mapped products (store_id: %s): %v", storeID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to remove mapped products"})
	}

	response := map[string]interface{}{
		"message":        "Mapped products processed",
		"deleted_skus":   deletedSKUs,
		"failed_skus":    failedSKUs,
	}

	return c.JSON(http.StatusOK, response)
}
