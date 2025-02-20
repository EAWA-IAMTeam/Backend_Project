package main

import (
	//"backend_project/database"
	"backend_project/database"
	"backend_project/internal/config"
	"backend_project/internal/payment/handlers"
	"backend_project/internal/payment/repositories"
	"backend_project/internal/payment/services"
	"backend_project/sdk"
	"fmt"

	//"fmt"
	"log"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

func main() {
	// // Connect to the database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	} else {
		fmt.Println("Connected to the database successfully!")
	}

	defer db.Close()

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

	//Define routes
	e := echo.New()

	returnRepo := repositories.NewReturnRepository(iopClient, env.AppKey, env.AccessToken, db)
	ordersRepo := repositories.NewOrdersRepository(iopClient, env.AppKey, env.AccessToken, db)
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
	ordersHandler := handlers.NewOrdersHandler(ordersService, itemListService, returnHandler, paymentService)

	//Testing
	// e.GET("/payments/orders", paymentHandler.GetOrders)
	// e.GET("/payments/transaction", paymentHandler.GetTransactions)
	// e.GET("/payments/payout", paymentHandler.GetPayouts)
	e.GET("/payments/transactionByOrder/:company_id", ordersHandler.GetTransactionsByOrder)
	// e.GET("/payments/transactionByPayout", paymentHandler.GetTransactionsByPayout)

	// e.GET("/payments/payment/order", paymentHandler.GetPaymentsByOrderID)

	// Start the server
	serverAddr := "192.168.0.184:8000"
	log.Printf("Server started at %s", serverAddr)
	if err := e.Start(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
