package main

import (
	"backend_project/internal/middleware/handlers"
	"log"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	// Initialize NATS connection
	if err := handlers.InitNATS(); err != nil {
		log.Fatal("Failed to initialize NATS:", err)
	}

	// Get NATS connection and JetStream context from handlers package
	nc := handlers.GetNATSConnection()
	js := handlers.GetJetStreamContext()

	// Initialize handlers
	requestHandler := handlers.NewRequestHandler(nc, js)

	// Setup routes
	// e.POST("/api/orders", requestHandler.HandlePostRequest)
	// /company/:company_id/employee/:employeeID
	e.GET("company/:company_id/topic/:topic", requestHandler.HandleGetRequest)
	e.GET("company/:company_id/topic/:topic/method/linkstore", requestHandler.LinkStore)
	e.GET("company/:company_id/topic/:topic/method/:method", requestHandler.GetStore)
	//e.GET("/callback", handleLazadaCallback)

	log.Println("🚀 API Gateway running on :8081")
	e.Logger.Fatal(e.Start(":8081"))
}
