package middleware

import (
	"bytes"
	"io"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

func CORSConfig() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
	})
}

// RequestLoggerMiddleware logs each API request with method, URL, and body (if applicable)
func RequestLoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		startTime := time.Now()

		// Read Request Body
		body, err := io.ReadAll(c.Request().Body)
		if err == nil && len(body) > 0 {
			// Restore the request body after reading (since ReadAll drains it)
			c.Request().Body = io.NopCloser(bytes.NewBuffer(body))
		}

		// Log the request
		log.Printf("[REQUEST] %s %s - Body: %s", c.Request().Method, c.Request().URL, string(body))

		// Call the next handler
		err = next(c)

		// Log the response time
		log.Printf("[RESPONSE] %s %s - Took %v", c.Request().Method, c.Request().URL, time.Since(startTime))

		return err
	}
}
