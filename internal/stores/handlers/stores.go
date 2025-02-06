package handlers

import (
	"backend_project/sdk"
	"database/sql"
	"log"
	"net/http"

	"github.com/labstack/echo"
)

type StoresHandler struct {
}

type AuthRequest struct {
	Code string `json:"code"`
}

// GetSellerInfo fetches seller information from Lazada API
func GenerateAccessToken(db *sql.DB, client *sdk.IopClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		authCode := c.QueryParam("code") // Get the authorization code from the request

		if authCode == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Authorization code is required",
			})
		}

		client.AddAPIParam("code", authCode) // Set the authorization code dynamically

		// Fetch seller details from Lazada API
		resp, err := client.Execute("/auth/token/create", "GET", nil)
		if err != nil {
			log.Printf("Failed to fetch seller info from Lazada API: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Failed to fetch seller information",
			})
		}

		// Print response to the console
		log.Printf("Lazada API Response: %+v", resp)

		return c.JSON(http.StatusOK, resp)
	}
}
