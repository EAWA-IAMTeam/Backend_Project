package main

import (
	"backend_project/internal/config"
	"backend_project/internal/orders/handlers"
	"backend_project/internal/orders/repositories"
	"backend_project/internal/orders/services"
	"backend_project/sdk"
	"database/sql"
	"log"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables from .env file
	env := config.LoadConfig()

	// Initialize SDK client with API credentials
	clientOptions := sdk.ClientOptions{
		APIKey:    env.AppKey,
		APISecret: env.AppSecret,
		Region:    "MY",
	}
	iopClient := sdk.NewClient(&clientOptions)
	iopClient.SetAccessToken(env.AccessToken)

	// Initialize Echo server
	e := echo.New()

	// Initialize PostgreSQL database connection
	dbConnStr := "user=" + env.DbUser + " dbname=" + env.DbName + " sslmode=disable"
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	defer db.Close()

	// Set up the repository, service, and handler
	ordersRepo := repositories.NewOrdersRepository(iopClient, env.AppKey, env.AccessToken)
	itemListRepo := repositories.NewItemListRepository(iopClient, env.AppKey, env.AccessToken)

	itemListService := services.NewItemListService(itemListRepo)
	ordersService := services.NewOrdersService(ordersRepo, itemListService)

	ordersHandler := handlers.NewOrdersHandler(ordersService, itemListService)

	// Define API routes
	e.GET("/orders", ordersHandler.GetOrders)

	// Start the server on IP 192.168.0.240 and port 8080
	serverAddr := "192.168.0.240:8080"
	log.Printf("Server started at %s", serverAddr)
	if err := e.Start(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
