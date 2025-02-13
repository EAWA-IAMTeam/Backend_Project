package main

import (
	"backend_project/database"
	"backend_project/internal/config"
	"backend_project/internal/stores/handlers"
	"backend_project/internal/stores/repositories"
	"backend_project/internal/stores/services"
	"backend_project/sdk"
	"log"

	"github.com/labstack/echo"
)

func main() {
	// Load configuration
	env := config.LoadConfig()

	// Connect to the database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// Initialize Lazada SDK client
	clientOptions := sdk.ClientOptions{
		APIKey:    env.AppKey,
		APISecret: env.AppSecret,
		Region:    "MY",
	}
	iopClient := sdk.NewClient(&clientOptions)

	// Initialize repository, service layers and handlers
	storeRepo := repositories.NewStoreRepository(db)
	storeService := services.NewStoreService(storeRepo, iopClient)
	storeHandler := handlers.NewStoreHandler(storeService)

	// Create a new echo instance
	e := echo.New()

	// Define routes and pass the service to handlers
	e.GET("/lazada/link/store", storeHandler.LazadaLinkStore)

	// Start the server
	e.Logger.Fatal(e.Start(":8100"))
}
