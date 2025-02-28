package handlers

import (
	//"backend_project/internal/stores/models"
	"backend_project/internal/stores/services"
	//"strconv"

	//"backend_project/internal/stores/repositories"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

type StoreHandler struct {
	storeService services.StoreService
	js           nats.JetStreamContext
}

func NewOrderHandler(ss services.StoreService, js nats.JetStreamContext) *StoreHandler {
	return &StoreHandler{
		storeService: ss,
		js:           js}
}

// SetupSubscriptions initializes all NATS subscriptions
func (h *StoreHandler) SetupSubscriptions() error {
	if _, err := h.js.QueueSubscribe("store.request.linkstore", "store-company", h.handleLazadaLinkStore); err != nil {
		return err
	}

	if _, err := h.js.QueueSubscribe("store.request.getstore", "store-getcompany", h.handleLazadaGetStore); err != nil {
		return err
	}

	log.Println("Store subscriptions setup complete")
	return nil
}

func (sh *StoreHandler) handleLazadaLinkStore(msg *nats.Msg) {

	var request struct {
		CompanyID int64  `json:"company_id"`
		RequestID string `json:"request_id"`
		Code      string `json:"code"`
	}

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		log.Println("Failed to unmarshal request:", err)
		sh.respondWithError("Invalid request format", request.RequestID)
		msg.Ack()
		return
	}

	authCode := request.Code
	companyIDStr := request.CompanyID
	requestID := request.RequestID

	if requestID == "" {
		sh.respondWithError("Request ID is required", requestID)
		msg.Ack() // Consider Nak() if retrying is needed
		return
	}

	if authCode == "" {
		sh.respondWithError("Authorization code is required", requestID)
		msg.Ack()
		return
	}

	if companyIDStr == 0 {
		sh.respondWithError("Company ID is required", requestID)
		msg.Ack()
		return
	}

	// Convert companyID to int64
	// companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	// if err != nil {
	// 	sh.respondWithError("Invalid Company ID format", requestID)
	// 	msg.Ack()
	// 	return
	// }

	response, err := sh.storeService.FetchStoreInfo(authCode, companyIDStr)
	if err != nil {
		sh.respondWithError(err.Error(), requestID)
		msg.Ack()
		return
	}

	// Send response using Jetstream
	responseData, err := json.Marshal(response)
	if err != nil {
		sh.respondWithError("Internal server error", requestID)
		msg.Ack()
		return
	}

	responseSubject := "store.response." + requestID
	if _, err := sh.js.Publish(responseSubject, responseData); err != nil {
		log.Printf("Error sending response: %v", err)
	}

	msg.Ack()
}

func (sh *StoreHandler) handleLazadaGetStore(msg *nats.Msg) {
	var request struct {
		CompanyID int64  `json:"company_id"`
		RequestID string `json:"request_id"`
	}

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		log.Println("Failed to unmarshal request:", err)
		sh.respondWithError("Invalid request format", request.RequestID)
		msg.Ack()
		return
	}

	// Validate request parameters
	if request.RequestID == "" {
		sh.respondWithError("Request ID is required", request.RequestID)
		msg.Ack()
		return
	}

	if request.CompanyID == 0 {
		sh.respondWithError("Company ID is required", request.RequestID)
		msg.Ack()
		return
	}

	// Fetch stores using the service
	stores, err := sh.storeService.GetStoresByCompany(request.CompanyID)
	if err != nil {
		sh.respondWithError(err.Error(), request.RequestID)
		msg.Ack()
		return
	}

	// Send the response
	responseData, err := json.Marshal(stores)
	if err != nil {
		sh.respondWithError("Internal server error", request.RequestID)
		msg.Ack()
		return
	}

	responseSubject := "store.response." + request.RequestID
	if _, err := sh.js.Publish(responseSubject, responseData); err != nil {
		log.Printf("Error sending response: %v", err)
	}

	msg.Ack()
}

func (sh *StoreHandler) respondWithError(errMsg, requestID string) {
	response := map[string]string{"error": errMsg}
	data, _ := json.Marshal(response)

	responseSubject := "store.response." + requestID
	sh.js.Publish(responseSubject, data)
}
