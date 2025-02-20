package handlers

import (
	"backend_project/internal/orders/repositories"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

type OrderHandler struct {
	repo *repositories.OrderRepository
	nc   *nats.Conn
}

func NewOrderHandler(repo *repositories.OrderRepository, nc *nats.Conn) *OrderHandler {
	return &OrderHandler{
		repo: repo,
		nc:   nc}
}

// SetupSubscriptions initializes all NATS subscriptions
func (h *OrderHandler) SetupSubscriptions() error {
	// Subscribe to get orders by company
	if _, err := h.nc.Subscribe("order.company.get", h.handleGetOrdersByCompany); err != nil {
		return err
	}

	log.Println("📦 Order subscriptions setup complete")
	return nil
}

func (oh *OrderHandler) handleGetOrdersByCompany(msg *nats.Msg) {
	//Extract comapny ID from subject
	var request struct {
		CompanyID int8 `json:"company_id"`
	}

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		log.Println("Failed to unmarshal request:", err)
		oh.respondWithError(msg, "Invalid request format")
		return
	}

	// Get orders from repository
	orders, err := oh.repo.GetOrdersByCompany(request.CompanyID)
	if err != nil {
		log.Printf("Error fetching orders: %v", err)
		oh.respondWithError(msg, "Failed to fetch orders")
		return
	}

	// Send response
	response, err := json.Marshal(orders)
	if err != nil {
		oh.respondWithError(msg, "Internal server error")
		return
	}

	if err := msg.Respond(response); err != nil {
		log.Printf("Error sending response: %v", err)
	}
}

func (h *OrderHandler) respondWithError(msg *nats.Msg, errMsg string) {
	response := map[string]string{"error": errMsg}
	data, _ := json.Marshal(response)
	msg.Respond(data)
}
