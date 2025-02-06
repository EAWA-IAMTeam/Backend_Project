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
	e.Use(middleware.RequestLoggerMiddleware) // 👈 Add API logging middleware

	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	// Define routes
	e.GET("/products/company/:company_id", productHandler.GetStockItemsByCompany)
	e.GET("/products/store/:store_id", productHandler.GetProductsByStore)
	e.POST("/products", productHandler.InsertProducts)
	e.GET("/products/lazada/mapped/:store_id", productHandler.GetMappedProducts)
	e.GET("/products/lazada/unmapped/:store_id", productHandler.GetUnmappedProducts)

	// Start server
	port := "7000"
	address := "192.168.0.73:" + port
	fmt.Println("Server running on", address)
	e.Logger.Fatal(e.Start(address))
}
