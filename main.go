package main

import (
	"backend_project/database"
	"backend_project/internal/config"
	"backend_project/internal/orders/handlers"
	"backend_project/internal/orders/repositories"
	"backend_project/internal/orders/services"
	"backend_project/sdk"
	"log"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

func main() {
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Load environment variables from .env file
	env := config.LoadConfig()

	// Initialize SDK client with API credentials
	clientOptions := sdk.ClientOptions{
		APIKey:    env.AppKey,
		APISecret: env.AppSecret,
		Region:    "MY",
	}
	iopClient := sdk.NewClient(&clientOptions)

	// Note: We no longer set a global access token here since it's retrieved
	// from the database per request based on company_id

	// Initialize Echo server
	e := echo.New()

	// Repositories
	returnRepo := repositories.NewReturnRepository(iopClient, env.AppKey, env.AccessToken, db)
	ordersRepo := repositories.NewOrdersRepository(iopClient, db, env.AppKey)
	itemListRepo := repositories.NewItemListRepository(iopClient, env.AppKey, env.AccessToken, db)
	paymentRepo := repositories.NewPaymentsRepository(iopClient, env.AppKey, env.AccessToken, db)

	// Services
	itemListService := services.NewItemListService(itemListRepo)
	returnService := services.NewReturnService(returnRepo)
	paymentService := services.NewPaymentService(paymentRepo)
	ordersService := services.NewOrdersService(ordersRepo, itemListService, returnService, paymentService)

	// Initialize return handler
	returnHandler := handlers.NewReturnHandler(returnService)

	// Initialize orders handler with all required services
	ordersHandler := handlers.NewOrdersHandler(ordersService, itemListService, returnHandler, paymentService, db)

	// Define API routes
	e.GET("/orders/:company_id", ordersHandler.GetOrders)
	e.GET("/orders/:company_id/:status", ordersHandler.GetOrders)
	e.GET("/orders/:company_id/E", ordersHandler.FetchOrdersByCompanyID)
	e.GET("/orders/:company_id/E1", ordersHandler.GetTransactionsByOrder)

	// Start the server on IP 192.168.0.240 and port 8080
	serverAddr := "192.168.0.240:8000"
	log.Printf("Server started at %s", serverAddr)
	if err := e.Start(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
