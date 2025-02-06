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
func (ph *ProductHandler) GetProductsByStore(c echo.Context) error {
	storeID, err := strconv.Atoi(c.Param("store_id"))
	if err != nil {
		log.Printf("Failed to convert store_id to int: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid store_id"})
	}

	products, err := ph.ProductService.FetchProductsByStore(storeID)
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

// GetFilteredProducts handles API requests for filtered products
func (ph *ProductHandler) GetFilteredProducts(c echo.Context) error {
	storeID := c.Param("store_id")
	if storeID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "store_id is required"})
	}

	products, err := ph.ProductService.FetchFilteredProducts(storeID)
	if err != nil {
		log.Printf("Error fetching filtered products: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to fetch products"})
	}

	return c.JSON(http.StatusOK, products)
}