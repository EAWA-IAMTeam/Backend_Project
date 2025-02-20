package handlers

import (
	"backend_project/internal/products/models"
	"backend_project/internal/products/services"
	"fmt"
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
	var companyID int64
	var err error

	companyID, err = strconv.ParseInt(c.Param("company_id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid company_id"})
	}

	stockItems, err := ph.ProductService.FetchStockItemsByCompany(companyID)
	if err != nil {
		log.Printf("Failed to fetch stock items: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to fetch stock items"})
	}

	return c.JSON(http.StatusOK, stockItems)
}

// PostStockItemsByCompany handles inserting stock items for a specific company
func (ph *ProductHandler) PostStockItemsByCompany(c echo.Context) error {
	var companyID int64
	var err error

	companyID, err = strconv.ParseInt(c.Param("company_id"), 10, 64)
	if err != nil {
		fmt.Println("Error parsing company_id:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid company_id"})
	}

	var request struct {
		StockItems []models.StockItem `json:"stock_items"`
	}

	if err := c.Bind(&request); err != nil {
		fmt.Println("Error binding request body:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
	}

	if len(request.StockItems) == 0 {
		fmt.Println("Error: Stock items list is empty")
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Stock items cannot be empty"})
	}

	err = ph.ProductService.CreateStockItemsByCompany(companyID, request.StockItems)
	if err != nil {
		fmt.Println("Error in CreateStockItemsByCompany:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to post stock items"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Stock items successfully posted"})
}

// GetProductsByStore handles fetching products by store ID
func (ph *ProductHandler) GetProductsByCompany(c echo.Context) error {
	var companyID int64
	var err error

	companyID, err = strconv.ParseInt(c.Param("company_id"), 10, 64)
	if err != nil {
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
	var companyID int64
	var err error

	companyID, err = strconv.ParseInt(c.Param("company_id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid company_id"})
	}

	products, err := ph.ProductService.FetchMappedProducts(companyID)
	if err != nil {
		log.Printf("Error fetching mapped products: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to fetch mapped products"})
	}

	return c.JSON(http.StatusOK, products)
}

// GetUnmappedProducts handles API requests for mapped products
// TODO: Fetch the products from all platforms according to the company's store by using the access token in database
func (ph *ProductHandler) GetUnmappedProducts(c echo.Context) error {
	var companyID int64
	var err error

	companyID, err = strconv.ParseInt(c.Param("company_id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid company_id"})
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
	var request struct {
		StoreID int64  `json:"store_id"`
		SKU     string `json:"sku"`
	}

	// Parse JSON body
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
	}

	rowsAffected, err := ph.ProductService.DeleteMappedProducts(request.StoreID, request.SKU)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to remove mapped product"})
	}

	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Product not found or already removed"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Mapped product successfully removed"})
}

// RemoveMappedProductsBatch handles API requests to delete multiple mapped products
func (ph *ProductHandler) RemoveMappedProductsBatch(c echo.Context) error {
	var request struct {
		StoreID int64    `json:"store_id"`
		SKUs    []string `json:"skus"`
	}

	// Parse JSON body
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
	}

	if len(request.SKUs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "At least one SKU is required"})
	}

	// Use request.StoreID instead of undefined storeID
	deletedSKUs, failedSKUs, err := ph.ProductService.DeleteMappedProductsBatch(request.StoreID, request.SKUs)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to remove mapped products"})
	}

	response := map[string]interface{}{
		"message":      "Mapped products processed",
		"deleted_skus": deletedSKUs,
		"failed_skus":  failedSKUs,
	}

	return c.JSON(http.StatusOK, response)
}
