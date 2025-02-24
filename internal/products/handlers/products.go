package handlers

import (
	"backend_project/internal/products/models"
	"backend_project/internal/products/services"
	"encoding/json"
	"log"
	"net/http"

	"github.com/nats-io/nats.go"
)

type ProductHandler struct {
	ps *services.ProductService
	js nats.JetStreamContext
}

func NewProductHandler(ps *services.ProductService, js nats.JetStreamContext) *ProductHandler {
	return &ProductHandler{
		ps: ps,
		js: js,
	}
}

// SetupSubscriptions initializes all NATS subscriptions
func (ph *ProductHandler) SetupSubscriptions() error {
	// Subscribe to get product by company
	if _, err := ph.js.QueueSubscribe("product.request.getsqlitembycompany", "product-sql-company", ph.GetSQLItemsByCompany); err != nil {
		return err
	}

	if _, err := ph.js.QueueSubscribe("product.request.postsqlitem", "product-sql-post", ph.PostSQLItemsByCompany); err != nil {
		return err
	}

	if _, err := ph.js.QueueSubscribe("product.request.getproductbycompany", "product-product-company", ph.GetProductsByCompany); err != nil {
		return err
	}

	if _, err := ph.js.QueueSubscribe("product.request.insertproducts", "product-insert-products", ph.InsertProducts); err != nil {
		return err
	}

	if _, err := ph.js.QueueSubscribe("product.request.getmappedproducts", "product-product-mapped", ph.GetMappedProducts); err != nil {
		return err
	}

	if _, err := ph.js.QueueSubscribe("product.request.getunmappedproducts", "product-product-unmapped", ph.GetUnmappedProducts); err != nil {
		return err
	}

	// Subscribe to delete product by SKU
	if _, err := ph.js.QueueSubscribe("product.request.deleteproduct", "product-delete-product", ph.RemoveMappedProducts); err != nil {
		return err
	}

	// Subscribe to delete products batch by SKUs
	if _, err := ph.js.QueueSubscribe("product.request.deleteproductsbatch", "product-delete-products-batch", ph.RemoveMappedProductsBatch); err != nil {
		return err
	}

	log.Println("Product subscriptions setup complete")
	return nil
}

// GetStockItemsByCompany handles fetching stock items by company ID
func (ph *ProductHandler) GetSQLItemsByCompany(msg *nats.Msg) {
	var request models.PaginatedRequest

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		log.Println("Failed to unmarshal request:", err)
		ph.respondWithError("Invalid request format", request.RequestID)
		msg.Ack()
		return
	}

	//Get stock Items from services
	sqlItems, err := ph.ps.FetchStockItemsByCompany(
		request.CompanyID,
		request.Pagination.Page,
		request.Pagination.Limit,
	)
	if err != nil {
		log.Printf("Error fetching sql items: %v", err)
		ph.respondWithError("Failed to fetch sql items", request.RequestID)
		msg.Ack()
		return
	}

	// Send response using Jetstream
	response, err := json.Marshal(sqlItems)
	if err != nil {
		ph.respondWithError("Internal server error", request.RequestID)
		msg.Ack()
		return
	}

	//Publish response to JetStream (`product.response.<requestID>`)
	responseSubject := "product.response." + request.RequestID
	if _, err := ph.js.Publish(responseSubject, response); err != nil {
		log.Printf("Error sending response: %v", err)
	}

	// Explicitly acknowledge message
	msg.Ack()
}

// PostStockItemsByCompany handles inserting stock items for a specific company
func (ph *ProductHandler) PostSQLItemsByCompany(msg *nats.Msg) {
	//Parse request payload
	var request struct {
		CompanyID  int64              `json:"company_id"`
		RequestID  string             `json:"request_id"`
		StockItems []models.StockItem `json:"stock_items"`
	}

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		log.Println("Failed to unmarshal request:", err)
		ph.respondWithError("Invalid request format", request.RequestID)
		msg.Ack()
		return
	}

	// Validate sql items
	if len(request.StockItems) == 0 {
		log.Println("Error: Stock items list is empty")
		ph.respondWithError("Stock items cannot be empty", request.RequestID)
		msg.Ack()
		return
	}

	err := ph.ps.CreateStockItemsByCompany(request.CompanyID, request.StockItems)
	if err != nil {
		log.Printf("Error inserting stock items: %v", err)
		ph.respondWithError("Failed to create stock items", request.RequestID)
		msg.Ack()
		return
	}

	//Prepare success response
	response := map[string]interface{}{
		"message":    "Stock items successfully created",
		"company_id": request.CompanyID,
		"request_id": request.RequestID,
	}

	// Send response back to JetStream (`product.response.{request_id}`)
	responseData, _ := json.Marshal(response)
	responseSubject := "product.response." + request.RequestID
	if _, err := ph.js.Publish(responseSubject, responseData); err != nil {
		log.Printf("Failed to publish response: %v", err)
	}

	msg.Ack()
	log.Printf("Successfully processed stock items for company %d", request.CompanyID)
}

// GetProductsByStore handles fetching products by store ID
func (ph *ProductHandler) GetProductsByCompany(msg *nats.Msg) {
	var request models.PaginatedRequest

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		log.Println("Failed to unmarshal request:", err)
		ph.respondWithError("Invalid request format", request.RequestID)
		msg.Ack()
		return
	}

	products, err := ph.ps.FetchProductsByCompany(
		request.CompanyID,
		request.Pagination.Page,
		request.Pagination.Limit,
	)

	if err != nil {
		log.Printf("Error fetching sql items: %v", err)
		ph.respondWithError("Failed to fetch sql items", request.RequestID)
		msg.Ack()
		return
	}

	// Send response using Jetstream
	response, err := json.Marshal(products)
	if err != nil {
		ph.respondWithError("Internal server error", request.RequestID)
		msg.Ack()
		return
	}

	//Publish response to JetStream (`product.response.<requestID>`)
	responseSubject := "product.response." + request.RequestID
	if _, err := ph.js.Publish(responseSubject, response); err != nil {
		log.Printf("Error sending response: %v", err)
	}

	// Explicitly acknowledge message
	msg.Ack()
}

// InsertProducts handles inserting products into the database
func (ph *ProductHandler) InsertProducts(msg *nats.Msg) {
	req, err := ph.ps.ParseProductRequest(msg)
	// Extract request_id from the first product in the slice (if available)
	var requestID string
	if len(req) > 0 {
		requestID = req[0].RequestID // Ensure StoreProduct has a RequestID field
	}

	if err != nil {
		log.Println("Failed to unmarshal request:", err)
		ph.respondWithError("Invalid request format", requestID)
		msg.Ack()
		return
	}

	result, err := ph.ps.InsertProducts(req)
	if err != nil {
		log.Println("Failed to unmarshal request:", err)
		ph.respondWithError("Invalid request format", requestID)
		msg.Ack()
		return
	}

	statusCode := http.StatusCreated
	if len(result.Duplicates) > 0 {
		statusCode = http.StatusConflict
	}

	// Construct response
	response := map[string]interface{}{
		"request_id": requestID,
		"status":     statusCode,
		"message":    "Products inserted successfully",
		"data":       result,
	}
	// Send response back to JetStream (`product.response.{request_id}`)
	responseData, _ := json.Marshal(response)
	responseSubject := "product.response." + requestID
	if _, err := ph.js.Publish(responseSubject, responseData); err != nil {
		log.Printf("Failed to publish response: %v", err)
	}

	msg.Ack()

}

// GetMappedProducts handles API requests for mapped products
func (ph *ProductHandler) GetMappedProducts(msg *nats.Msg) {
	var request models.PaginatedRequest

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		log.Println("Failed to unmarshal request:", err)
		ph.respondWithError("Invalid request format", request.RequestID)
		msg.Ack()
		return
	}

	products, err := ph.ps.FetchMappedProducts(
		request.CompanyID,
	)
	if err != nil {
		log.Printf("Error fetching mapped products: %v", err)
		ph.respondWithError("Failed to fetch mapped products", request.RequestID)
		msg.Ack()
		return
	}

	// Send response using Jetstream
	response, err := json.Marshal(products)
	if err != nil {
		ph.respondWithError("Internal server error", request.RequestID)
		msg.Ack()
		return
	}

	//Publish response to JetStream (`product.response.<requestID>`)
	responseSubject := "product.response." + request.RequestID
	if _, err := ph.js.Publish(responseSubject, response); err != nil {
		log.Printf("Error sending response: %v", err)
	}

	// Explicitly acknowledge message
	msg.Ack()
}

// GetUnmappedProducts handles API requests for mapped products
// TODO: Fetch the products from all platforms according to the company's store by using the access token in database
func (ph *ProductHandler) GetUnmappedProducts(msg *nats.Msg) {
	var request models.PaginatedRequest

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		log.Println("Failed to unmarshal request:", err)
		ph.respondWithError("Invalid request format", request.RequestID)
		msg.Ack()
		return
	}

	products, err := ph.ps.FetchUnmappedProducts(request.CompanyID)
	if err != nil {
		log.Printf("Error fetching unmapped products: %v", err)
		ph.respondWithError("Failed to fetch unmapped products", request.RequestID)
		msg.Ack()
		return
	}

	// Send response using Jetstream
	response, err := json.Marshal(products)
	if err != nil {
		ph.respondWithError("Internal server error", request.RequestID)
		msg.Ack()
		return
	}

	//Publish response to JetStream (`product.response.<requestID>`)
	responseSubject := "product.response." + request.RequestID
	if _, err := ph.js.Publish(responseSubject, response); err != nil {
		log.Printf("Error sending response: %v", err)
	}

	// Explicitly acknowledge message
	msg.Ack()
}

// RemoveMappedProducts handles API requests to delete mapped products
func (ph *ProductHandler) RemoveMappedProducts(msg *nats.Msg) {
	var request struct {
		StoreID   int64  `json:"store_id"`
		SKU       string `json:"sku"`
		RequestID string `json:"request_id"`
	}

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		log.Println("Failed to unmarshal request:", err)
		ph.respondWithError("Invalid request format", "")
		msg.Ack()
		return
	}

	rowsAffected, err := ph.ps.DeleteMappedProducts(request.StoreID, request.SKU)
	if err != nil {
		log.Printf("Failed to delete product: %v", err)
		ph.respondWithError("Failed to delete product", request.RequestID)
		msg.Ack()
		return
	}

	if rowsAffected == 0 {
		ph.respondWithError("Product not found or already deleted", request.RequestID)
		msg.Ack()
		return
	}

	// Prepare success response
	response := map[string]interface{}{
		"message":    "Product successfully deleted",
		"store_id":   request.StoreID,
		"sku":        request.SKU,
		"request_id": request.RequestID,
	}

	// Send response back to JetStream (`product.response.{request_id}`)
	responseData, _ := json.Marshal(response)
	responseSubject := "product.response." + request.RequestID
	log.Println(responseSubject)
	log.Println(responseData)
	if _, err := ph.js.Publish(responseSubject, responseData); err != nil {
		log.Printf("Failed to publish response: %v", err)
	}

	msg.Ack()
}

// RemoveMappedProductsBatch handles API requests to delete multiple mapped products
func (ph *ProductHandler) RemoveMappedProductsBatch(msg *nats.Msg) {
	var request struct {
		StoreID   int64    `json:"store_id"`
		SKUs      []string `json:"skus"`
		RequestID string   `json:"request_id"`
	}

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		log.Println("Failed to unmarshal request:", err)
		ph.respondWithError("Invalid request format", "")
		msg.Ack()
		return
	}

	if len(request.SKUs) == 0 {
		ph.respondWithError("At least one SKU is required", request.RequestID)
		msg.Ack()
		return
	}

	deletedSKUs, failedSKUs, err := ph.ps.DeleteMappedProductsBatch(request.StoreID, request.SKUs)
	if err != nil {
		log.Printf("Failed to delete products: %v", err)
		ph.respondWithError("Failed to delete products", request.RequestID)
		msg.Ack()
		return
	}

	// Prepare success response
	response := map[string]interface{}{
		"message":      "Mapped products processed",
		"deleted_skus": deletedSKUs,
		"failed_skus":  failedSKUs,
		"request_id":   request.RequestID,
	}

	// Send response back to JetStream (`product.response.{request_id}`)
	responseData, _ := json.Marshal(response)
	responseSubject := "product.response." + request.RequestID
	if _, err := ph.js.Publish(responseSubject, responseData); err != nil {
		log.Printf("Failed to publish response: %v", err)
	}

	msg.Ack()
}

func (ph *ProductHandler) respondWithError(errMsg string, requestID string) {
	response := map[string]string{"error": errMsg}
	data, _ := json.Marshal(response)

	responseSubject := "product.response." + requestID
	ph.js.Publish(responseSubject, data)
}
