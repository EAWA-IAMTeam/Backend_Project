package main

import (
	"backend_project/internal/middleware/handlers"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // Allow all origins (or restrict it to your frontend URL)
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}))

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
	e.GET("company/:company_id/topic/:topic/method/:method", requestHandler.HandleGetRequest)
	e.GET("/company/:company_id/topic/:topic/linkstore", requestHandler.LinkStore)
	e.GET("/company/:company_id/topic/:topic/getstore", requestHandler.GetStore)
	e.POST("company/:company_id/topic/:topic/method/postsqlitem", requestHandler.PostSQLItems)
	e.POST("company/:company_id/topic/:topic/method/insertproducts", requestHandler.PostProducts)
	e.DELETE("company/:company_id/topic/:topic/method/deleteproduct", requestHandler.DeleteProduct)
	e.DELETE("company/:company_id/topic/:topic/method/deleteproductsbatch", requestHandler.DeleteProductsBatch)

	log.Println("🚀 API Gateway running on :8081")
	e.Logger.Fatal(e.Start(":8081"))
}
