package main

import (
	"backend_project/database"
	"backend_project/internal/stores/handlers"
	"backend_project/internal/stores/repositories"
	"backend_project/internal/stores/services"
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
	repo := repositories.NewStoreRepository(db)
	storeService := services.NewStoreService(repo)
	handler := handlers.NewOrderHandler(storeService, js)

	// Setup NATS subscriptions
	if err := handler.SetupSubscriptions(); err != nil {
		log.Fatalf("Failed to setup subscriptions: %v", err)
	}

	log.Println("Store Service running...")
	select {} // Keep the service running
}
