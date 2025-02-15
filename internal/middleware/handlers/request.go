package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"backend_project/internal/middleware/models"

	"github.com/labstack/echo/v4"
)

//Handle Post Requests (Publish Event)

func HandlePostRequest(c echo.Context) error {
	req := new(models.RequestPayload)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
	}

	//Convert data to JSON
	data, _ := json.Marshal(req.Data)

	//Publish to JetStream
	_, err := js.Publish(req.Topic, data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to publish message"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Message published successfully"})
}

// Handle Get Requests (Consume Event)
func HandleGetRequest(c echo.Context) error {
	topic := c.Param("topic")           //Extract topic from query
	orderID := c.QueryParam("order_id") // Extract order ID

	if topic == "" || orderID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing required parameters"})
	}

	requestData := map[string]string{"order_id": orderID}
	data, _ := json.Marshal(requestData)

	//Request data from NATS (Order Status)
	msg, err := nc.Request(topic, data, 2*time.Second)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Timeout waiting for response"})
	}

	return c.JSON(http.StatusOK, json.RawMessage(msg.Data))
}
