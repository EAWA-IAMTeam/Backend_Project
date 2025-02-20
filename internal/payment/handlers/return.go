package handlers

import (
	"backend_project/internal/payment/services"
)

type ReturnHandler struct {
	service services.ReturnService
}

func NewReturnHandler(service services.ReturnService) *ReturnHandler {
	return &ReturnHandler{service}
}

func (h *ReturnHandler) HandleReturnRequest(orderID int) error {
	return nil
}