package main

import (
	"backend_project/database"
	"backend_project/internal/middleware"
	"backend_project/internal/products/handlers"
	"backend_project/internal/products/repositories"
	"backend_project/internal/products/services"
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

func main() {
	// Connect to the database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	} else {
		fmt.Println("Connected to the database successfully!")
	}

	// Create a new Echo instance
	e := echo.New()

	// Apply middleware
	e.Use(middleware.CORSConfig())
	e.Use(middleware.RequestLoggerMiddleware) // API logging middleware

	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	// Define routes
	e.GET("/products/stock-item/:company_id", productHandler.GetStockItemsByCompany)
	e.POST("/products/stock-item/:company_id", productHandler.PostStockItemsByCompany)
	e.GET("/products/store-products/:company_id", productHandler.GetProductsByCompany)
	e.POST("/products/store-products", productHandler.InsertProducts)
	e.GET("/products/mapped-products/:company_id", productHandler.GetMappedProducts)
	e.DELETE("products/mapped-product", productHandler.RemoveMappedProducts)
	e.DELETE("products/mapped-products", productHandler.RemoveMappedProductsBatch)

	// TODO: Fetch the products from all platforms according to the company's store by using the access token in database
	e.GET("/products/unmapped-products/:company_id", productHandler.GetUnmappedProducts)

	// Start server
	port := "7000"
	address := "192.168.0.73:" + port
	fmt.Println("Server running on", address)
	e.Logger.Fatal(e.Start(address))
}
