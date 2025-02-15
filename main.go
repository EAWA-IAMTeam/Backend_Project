package middleware

import (
	"backend_project/database"
	"backend_project/internal/middleware/handlers"
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
)

func main() {
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

	e := echo.New()

	e.POST("/api", handlers.HandlePostRequest)
	e.GET("/api", handlers.HandleGetRequest)

	// Start server
	log.Println("🚀 Middleware running on :8081")
	e.Logger.Fatal(e.Start(":8081"))
}
