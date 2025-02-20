package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"
)

// Handle Post Requests (Publish Event)
type RequestHandler struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func NewRequestHandler(nc *nats.Conn, js nats.JetStreamContext) *RequestHandler {
	return &RequestHandler{
		nc: nc,
		js: js,
	}
}

// Handle Get Requests (Consume Event)
func (h *RequestHandler) HandleGetRequest(c echo.Context) error {
	companyIDstr := c.QueryParam("company_id")
	companyID, err := strconv.Atoi(companyIDstr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid order_id"})
	}

	// Prepare request data
	request := map[string]int{"company_id": companyID}
	data, _ := json.Marshal(request)

	// Request data from Orders service
	msg, err := h.nc.Request("order.company.get", data, 5*time.Second)
	if err != nil {
		log.Printf("Error requesting orders: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch orders",
		})
	}

	// Parse and return response
	var response json.RawMessage
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Invalid response format",
		})
	}

	return c.JSON(http.StatusOK, response)
}
