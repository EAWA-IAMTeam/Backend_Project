package handlers

import (
	"backend_project/internal/orders/models"
	"backend_project/internal/orders/repositories"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

type OrderHandler struct {
	repo *repositories.OrderRepository
	js   nats.JetStreamContext
}

func NewOrderHandler(repo *repositories.OrderRepository, js nats.JetStreamContext) *OrderHandler {
	return &OrderHandler{
		repo: repo,
		js:   js}
}

// SetupSubscriptions initializes all NATS subscriptions
func (h *OrderHandler) SetupSubscriptions() error {
	// Subscribe to get orders by company
	if _, err := h.js.QueueSubscribe("order.request", "order-workers", h.handleGetOrdersByCompany); err != nil {
		return err
	}

	log.Println("Order subscriptions setup complete")
	return nil
}

func (oh *OrderHandler) handleGetOrdersByCompany(msg *nats.Msg) {
	//Extract comapny ID from subject
	var request models.PaginatedRequest

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		log.Println("Failed to unmarshal request:", err)
		oh.respondWithError("Invalid request format", request.RequestID)
		return
	}

	// Get orders from repository
	orders, _, err := oh.repo.GetOrdersByCompany(
		request.CompanyID,
		request.Pagination.Page,
		request.Pagination.Limit,
	)
	if err != nil {
		log.Printf("Error fetching orders: %v", err)
		oh.respondWithError("Failed to fetch orders", request.RequestID)
		return
	}

	// Send response using Jetstream
	response, err := json.Marshal(orders)
	if err != nil {
		oh.respondWithError("Internal server error", request.RequestID)
		return
	}

	//Publish response to JetStream (`Orders.response.<requestID>`)
	responseSubject := "order.response." + request.RequestID
	if _, err := oh.js.Publish(responseSubject, response); err != nil {
		log.Printf("Error sending response: %v", err)
	}

	// Explicitly acknowledge message
	msg.Ack()
}

func (h *OrderHandler) respondWithError(errMsg string, requestID string) {
	response := map[string]string{"error": errMsg}
	data, _ := json.Marshal(response)

	responseSubject := "order.response." + requestID
	h.js.Publish(responseSubject, data)
}
