package main

import (
	"backend_project/database"
	"backend_project/internal/config"
	"backend_project/internal/orders/handlers"
	"backend_project/internal/orders/repositories"
	"backend_project/internal/orders/services"
	"backend_project/sdk"
	"log"

	"github.com/nats-io/nats.go"
)

func main() {
	// Connect to database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
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

	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to initialize JetStream: %v", err)
	}

	// Initialize dependencies
	// Repositories
	returnRepo := repositories.NewReturnRepository(iopClient, env.AppKey, env.AccessToken, db)
	ordersRepo := repositories.NewOrderRepository(iopClient, env.AppKey, env.AccessToken, db)
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
	ordersHandler := handlers.NewOrdersHandler(js, ordersService, itemListService, returnHandler, paymentService)

	// Setup NATS subscriptions
	if err := ordersHandler.SetupSubscriptions(); err != nil {
		log.Fatalf("Failed to setup subscriptions: %v", err)
	}

	log.Println("Order Service running...")
	select {} // Keep the service running
}
