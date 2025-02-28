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
	// Ensure the ORDERS stream exists
	_, err := js.AddStream(&nats.StreamConfig{
		Name:     "order",
		Subjects: []string{"order.request.*", "order.response.*"},
		// MaxBytes:          370000000000,
		// MaxMsgSize:        370000000,
		// MaxMsgsPerSubject: 37000000,
		// MaxMsgs:           370000,
		Storage: nats.FileStorage, // Persistent storage
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		log.Fatalf("Failed to create JetStream stream: %v", err)
	}

	// Create a durable consumer for responses
	_, err = js.AddConsumer("order", &nats.ConsumerConfig{
		Durable:       "order-replies",
		FilterSubject: "order.response.*", // Listen to all responses
		AckPolicy:     nats.AckExplicitPolicy,
	})
	if err != nil && err != nats.ErrConsumerNameAlreadyInUse {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	log.Println("JetStream order stream initialized")
	return &RequestHandler{nc: nc, js: js}
}

// Handle Get Requests (Consume Event)
func (h *RequestHandler) HandleGetRequest(c echo.Context) error {
	companyID, _ := strconv.Atoi(c.Param("company_id"))
	topic := c.Param("topic")
	method := c.Param("method")

	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	status := c.QueryParam("status")
	createdAfter := c.QueryParam("created_after")
	stopAfter := c.QueryParam("stop_after")
	sortDirection := c.QueryParam("sort_direction")

	//Set defaults if not provided
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20 //default page size
	}

	// Generate a unique request ID
	requestID := "req-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	// Create request payload
	request := map[string]interface{}{
		"company_id":     companyID,
		"request_id":     requestID,
		"status":         status,
		"created_after":  createdAfter,
		"stop_after":     stopAfter,
		"sort_direction": sortDirection,
		"pagination": map[string]int{
			"page":  page,
			"limit": limit,
		},
	}
	data, _ := json.Marshal(request)

	//Subscribe to Jetstream to fetch messages
	_, err := h.js.Publish(topic+".request."+method, data)
	if err != nil {
		log.Printf("Failed to publish request: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to publish request"})
	}

	//Subscribe to shared response consumer (order.response.*)
	sub, err := h.js.PullSubscribe(topic+".response.*", topic+"-replies")
	if err != nil {
		log.Printf("Failed to subscribe for response: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to subscribe for response"})
	}
	// Fetch response (wait up to 5 seconds)
	timeout := time.After(5 * time.Minute)
	for {
		select {
		case <-timeout:
			return c.JSON(http.StatusGatewayTimeout, map[string]string{"error": "Timeout waiting for response"})
		default:
			msgs, err := sub.Fetch(1)
			if err != nil || len(msgs) == 0 {
				continue
			}
			msgs[0].Ack()

			// Parse and return response
			var response json.RawMessage
			if err := json.Unmarshal(msgs[0].Data, &response); err != nil {
				log.Printf("Error : %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid response format"})

			}

			// Acknowledge message processing
			msgs[0].Ack()

			return c.JSON(http.StatusOK, response)
		}
	}
}

// link store
func (h *RequestHandler) LinkStore(c echo.Context) error {
	companyID, _ := strconv.Atoi(c.Param("company_id"))
	topic := c.Param("topic")

	// page, _ := strconv.Atoi(c.QueryParam("page"))
	// limit, _ := strconv.Atoi(c.QueryParam("limit"))
	code := c.QueryParam("code")

	// Validate auth_code
	if code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Authorization code is required"})
	}

	// Generate a unique request ID
	requestID := "req-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	// Create request payload
	request := map[string]interface{}{
		"company_id": companyID,
		"code":       code,
		"request_id": requestID,
	}
	data, _ := json.Marshal(request)

	//Subscribe to Jetstream to fetch messages
	_, err := h.js.Publish(topic+".request.linkstore", data)
	if err != nil {
		log.Printf("Failed to publish request: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to publish request"})
	}

	//Subscribe to shared response consumer (order.response.*)
	sub, err := h.js.PullSubscribe(topic+".response.*", topic+"-replies")
	if err != nil {
		log.Printf("Failed to subscribe for response: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to subscribe for response"})
	}
	// Fetch response (wait up to 15 seconds)
	timeout := time.After(15 * time.Second)
	for {
		select {
		case <-timeout:
			return c.JSON(http.StatusGatewayTimeout, map[string]string{"error": "Timeout waiting for response"})
		default:
			msgs, err := sub.Fetch(1) // Set max wait per fetch
			if err != nil {
				log.Println(msgs)
				log.Printf("Failed to fetch message: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch response"})
			}

			if len(msgs) > 0 {
				// Process message
				var response json.RawMessage
				if err := json.Unmarshal(msgs[0].Data, &response); err != nil {
					log.Printf("Error parsing response: %v", err)
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid response format"})
				}

				msgs[0].Ack() // Acknowledge the message
				return c.JSON(http.StatusOK, response)
			}
		}
	}
}

func (h *RequestHandler) GetStore(c echo.Context) error {
	companyID, _ := strconv.Atoi(c.Param("company_id"))
	topic := c.Param("topic")

	// Generate a unique request ID
	requestID := "req-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	// Create request payload
	request := map[string]interface{}{
		"company_id": companyID,
		"request_id": requestID,
	}
	data, _ := json.Marshal(request)

	//Subscribe to Jetstream to fetch messages
	_, err := h.js.Publish(topic+".request.getstore", data)
	if err != nil {
		log.Printf("Failed to publish request: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to publish request"})
	}

	//Subscribe to shared response consumer (order.response.*)
	sub, err := h.js.PullSubscribe(topic+".response.*", topic+"-replies")
	if err != nil {
		log.Printf("Failed to subscribe for response: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to subscribe for response"})
	}
	// Fetch response (wait up to 15 seconds)
	timeout := time.After(15 * time.Second)
	for {
		select {
		case <-timeout:
			return c.JSON(http.StatusGatewayTimeout, map[string]string{"error": "Timeout waiting for response"})
		default:
			msgs, err := sub.Fetch(1) // Set max wait per fetch
			if err != nil {
				log.Println(msgs)
				log.Printf("Failed to fetch message: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch response"})
			}

			if len(msgs) > 0 {
				// Process message
				var response json.RawMessage
				if err := json.Unmarshal(msgs[0].Data, &response); err != nil {
					log.Printf("Error parsing response: %v", err)
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid response format"})
				}

				msgs[0].Ack() // Acknowledge the message
				return c.JSON(http.StatusOK, response)
			}
		}
	}
}

func (h *RequestHandler) PostSQLItems(c echo.Context) error {
	companyID, _ := strconv.Atoi(c.Param("company_id"))
	topic := c.Param("topic")

	//Read request body
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		log.Println(payload)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid response payload"})

	}

	//Generate a unique request ID
	requestID := "req-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	// Attach metadata to the request
	payload["company_id"] = companyID
	payload["request_id"] = requestID
	data, _ := json.Marshal(payload)

	//Publish request to Jetstream
	_, err := h.js.Publish(topic+".request.postsqlitem", data)
	if err != nil {
		log.Printf("Failed to publish request: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to publish request"})
	}

	//Subscribe for response
	responseSubject := topic + ".response.*"
	sub, err := h.js.PullSubscribe(responseSubject, topic+"-replies")
	if err != nil {
		log.Printf("Failed to subscribe for response : %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to subscribe for response"})
	}

	timeout := time.After(5 * time.Minute)
	for {
		select {
		case <-timeout:
			return c.JSON(http.StatusGatewayTimeout, map[string]string{"error": "Timeout waiting for response"})

		default:
			msgs, err := sub.Fetch(1)
			if err != nil || len(msgs) == 0 {
				continue
			}
			msgs[0].Ack()

			//Parse and return response
			var response json.RawMessage
			if err := json.Unmarshal(msgs[0].Data, &response); err != nil {
				log.Printf("Error parsing response: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid response format"})
			}

			//Acknowledge message processing
			msgs[0].Ack()

			return c.JSON(http.StatusOK, response)
		}

	}

}

func (h *RequestHandler) PostProducts(c echo.Context) error {
	companyID, _ := strconv.Atoi(c.Param("company_id"))
	topic := c.Param("topic")

	//Read request body
	var payload []map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		log.Println(payload)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid response payload"})

	}

	//Generate a unique request ID
	requestID := "req-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	// Attach metadata to each item in the payload
	for i := range payload {
		payload[i]["company_id"] = companyID
		payload[i]["request_id"] = requestID
	}

	data, _ := json.Marshal(payload)

	//Publish request to Jetstream
	_, err := h.js.Publish(topic+".request.insertproducts", data)
	if err != nil {
		log.Printf("Failed to publish request: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to publish request"})
	}

	//Subscribe for response
	responseSubject := topic + ".response.*"
	sub, err := h.js.PullSubscribe(responseSubject, topic+"-replies")
	if err != nil {
		log.Printf("Failed to subscribe for response : %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to subscribe for response"})
	}

	timeout := time.After(5 * time.Minute)
	for {
		select {
		case <-timeout:
			return c.JSON(http.StatusGatewayTimeout, map[string]string{"error": "aTimeout waiting for response"})

		default:
			msgs, err := sub.Fetch(1)
			if err != nil || len(msgs) == 0 {
				continue
			}
			msgs[0].Ack()

			//Parse and return response
			var response json.RawMessage
			if err := json.Unmarshal(msgs[0].Data, &response); err != nil {
				log.Printf("Error parsing response: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid response format"})
			}

			//Acknowledge message processing
			msgs[0].Ack()

			return c.JSON(http.StatusOK, response)
		}

	}

}

func (h *RequestHandler) DeleteProduct(c echo.Context) error {
	companyID, _ := strconv.Atoi(c.Param("company_id"))
	topic := c.Param("topic")

	// Read request body
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		log.Println(payload)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Generate a unique request ID
	requestID := "req-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	// Attach metadata to the request
	payload["company_id"] = companyID
	payload["request_id"] = requestID
	data, _ := json.Marshal(payload)

	// Publish request to JetStream
	_, err := h.js.Publish(topic+".request.deleteproduct", data)
	if err != nil {
		log.Printf("Failed to publish request: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to publish request"})
	}

	// Subscribe for response
	responseSubject := topic + ".response.*"
	sub, err := h.js.PullSubscribe(responseSubject, topic+"-replies")
	if err != nil {
		log.Println(responseSubject)
		log.Println(topic + requestID)
		log.Printf("Failed to subscribe for response: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to subscribe for response"})
	}

	timeout := time.After(5 * time.Minute)
	for {
		select {
		case <-timeout:
			return c.JSON(http.StatusGatewayTimeout, map[string]string{"error": "Timeout waiting for response"})
		default:
			msgs, err := sub.Fetch(1)
			if err != nil || len(msgs) == 0 {
				continue
			}
			msgs[0].Ack()

			// Parse and return response
			var response json.RawMessage
			if err := json.Unmarshal(msgs[0].Data, &response); err != nil {
				log.Printf("Error parsing response: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid response format"})
			}

			// Acknowledge message processing
			msgs[0].Ack()

			return c.JSON(http.StatusOK, response)
		}
	}
}

func (h *RequestHandler) DeleteProductsBatch(c echo.Context) error {
	companyID, _ := strconv.Atoi(c.Param("company_id"))
	topic := c.Param("topic")

	// Read request body
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		log.Println(payload)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Generate a unique request ID
	requestID := "req-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	// Attach metadata to the request
	payload["company_id"] = companyID
	payload["request_id"] = requestID
	data, _ := json.Marshal(payload)

	// Publish request to JetStream
	_, err := h.js.Publish(topic+".request.deleteproductsbatch", data)
	if err != nil {
		log.Printf("Failed to publish request: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to publish request"})
	}

	// Subscribe for response
	responseSubject := topic + ".response.*"
	sub, err := h.js.PullSubscribe(responseSubject, topic+"-replies")
	log.Println(responseSubject)
	if err != nil {
		// log.Printf("Failed to subscribe for response: %v", err)
		log.Println(responseSubject)
		log.Println(topic + requestID)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to subscribe for response"})
	}

	timeout := time.After(1 * time.Minute)
	for {
		select {
		case <-timeout:
			return c.JSON(http.StatusGatewayTimeout, map[string]string{"error": "Timeout waiting for response"})
		default:
			msgs, err := sub.Fetch(1)
			if err != nil || len(msgs) == 0 {
				continue
			}
			msgs[0].Ack()

			// Parse and return response
			var response json.RawMessage
			if err := json.Unmarshal(msgs[0].Data, &response); err != nil {
				log.Printf("Error parsing response: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid response format"})
			}

			// Acknowledge message processing
			msgs[0].Ack()

			return c.JSON(http.StatusOK, response)
		}
	}
}
