package main

import (
	"backend_project/database"
	"backend_project/internal/config"
	"backend_project/internal/stores/handlers"
	"backend_project/sdk"
	"fmt"
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

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	} else {
		fmt.Println("Connected to the database successfully!")
	}

	defer db.Close()

	// Initialize IOP client
	clientOptions := sdk.ClientOptions{
		APIKey:    env.AppKey,
		APISecret: env.AppSecret,
		Region:    "MY",
	}
	iopClient := sdk.NewClient(&clientOptions)
	// iopClient.SetAccessToken(env.AccessToken)

	// Create a new echo instance
	e := echo.New()

	e.GET("/lazada/link/store", handlers.LazadaGenerateAccessToken(db, iopClient))
	e.GET("/lazada/products", handlers.GetProducts(db, iopClient))

	// Start the server
	e.Logger.Fatal(e.Start(":8100"))

}
