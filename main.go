package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// API endpoints to test
var apiEndpoints = []string{
	// "http://192.168.0.240:8000/orders/2/E",
	// "http://192.168.11.136:8000/orders/2/E",
	// "http://192.168.0.189:8081/company/2/topic/order/method/getbycompany?page=1&limit=200",
	"http://192.168.11.27:8081/company/2/topic/product/method/getsqlitembycompany?page=1&limit=2",
	//"http://192.168.11.136:8000/products/stock-item/2",
}

// Number of concurrent requests per endpoint
const concurrency = 300

// Total requests per endpoint
const totalRequests = 5000

func sendRequest(wg *sync.WaitGroup, id int, url string) {
	defer wg.Done()

	resp, err := http.Get(url) // Use http.Post if needed
	if err != nil {
		fmt.Printf("[Request %d to %s] ❌ Error: %s\n", id, url, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[Request %d to %s] ❌ Failed to read response: %s\n", id, url, err)
		return
	}

	// Check if response is empty (NULL)
	if len(body) == 0 {
		fmt.Printf("[Request %d to %s] ⚠️ NULL RESPONSE RECEIVED!\n", id, url)
	} else {
		fmt.Printf("[Request %d to %s] ✅ Status: %d, Response: %s\n", id, url, resp.StatusCode, string(body))
	}
}

func spamEndpoint(url string, wg *sync.WaitGroup) {
	defer wg.Done()

	var innerWg sync.WaitGroup
	sem := make(chan struct{}, concurrency) // Limit concurrency per endpoint

	for i := 1; i <= totalRequests; i++ {
		innerWg.Add(1)
		sem <- struct{}{} // Acquire a slot
		go func(id int) {
			defer func() { <-sem }() // Release the slot
			sendRequest(&innerWg, id, url)
		}(i)
	}

	innerWg.Wait()
}

func main() {
	startTime := time.Now()
	var wg sync.WaitGroup

	// Start spamming all endpoints
	for _, url := range apiEndpoints {
		wg.Add(1)
		go spamEndpoint(url, &wg)
	}

	wg.Wait()
	fmt.Println("Completed all requests in", time.Since(startTime))
}
