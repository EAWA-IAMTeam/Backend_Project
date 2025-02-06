package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetUnmappedProducts handles API requests for unmapped products
func (ph *ProductHandler) GetUnmappedProducts(c echo.Context) error {
	storeID := c.Param("store_id")
	if storeID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "store_id is required"})
	}

	products, err := ph.ProductService.FetchUnmappedProducts(storeID)
	if err != nil {
		log.Printf("Error fetching unmapped products: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to fetch unmapped products"})
	}

	return c.JSON(http.StatusOK, products)
}

// GetMappedProducts handles API requests for mapped products
func (ph *ProductHandler) GetMappedProducts(c echo.Context) error {
	storeID := c.Param("store_id")
	if storeID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "store_id is required"})
	}

	products, err := ph.ProductService.FetchMappedProducts(storeID)
	if err != nil {
		log.Printf("Error fetching mapped products: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to fetch mapped products"})
	}

	return c.JSON(http.StatusOK, products)
}