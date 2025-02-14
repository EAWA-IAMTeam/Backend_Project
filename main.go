package main

import (
	"backend_project/database"
	"backend_project/internal/stores/handlers"
	"backend_project/internal/stores/repositories"
	"backend_project/internal/stores/services"
	"log"

	"github.com/labstack/echo"
)

func main() {

	// Connect to the database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// Initialize repository, service layers and handlers
	storeRepo := repositories.NewStoreRepository(db)
	storeService := services.NewStoreService(storeRepo)
	storeHandler := handlers.NewStoreHandler(storeService)

	// Create a new echo instance
	e := echo.New()

	// Define routes and pass the service to handlers
	e.GET("/lazada/link/store", storeHandler.LazadaLinkStore)

	// Start the server
	e.Logger.Fatal(e.Start(":8100"))
}
