package handlers

import (
	"backend_project/internal/stores/models"
	"backend_project/sdk"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/labstack/echo"
)

// GetSellerInfo fetches seller information from Lazada API
func LazadaGenerateAccessToken(db *sql.DB, client *sdk.IopClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		authCode := c.QueryParam("code")
		if authCode == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "Authorization code is required"})
		}

		client.AddAPIParam("code", authCode)

		// Call Lazada API to get the access token
		_, authResp, err := client.Execute("/auth/token/create", "GET", nil)
		if err != nil {
			log.Printf("Failed to fetch access token from Lazada API: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to fetch access token"})
		}

		// Parse response into struct
		var tokenData models.ApiResponseAccessToken
		jsonData, _ := json.Marshal(authResp) // Convert map to JSON
		if err := json.Unmarshal(jsonData, &tokenData); err != nil {
			log.Printf("Failed to parse access token response: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to parse access token response"})
		}

		if len(tokenData.UserInfo) == 0 {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "No user info found"})
		}

		userInfo := tokenData.UserInfo[0]
		isMainAccount := userInfo.UserID == userInfo.SellerID

		client.SetAccessToken(tokenData.AccessToken)

		// Check if seller exists in the database
		var storeID int64
		query := `SELECT id FROM store WHERE id = $1`
		err = db.QueryRow(query, userInfo.SellerID).Scan(&storeID)

		if err == sql.ErrNoRows {
			// If seller_id does not exist, fetch store details from Lazada
			resp, _, err := client.Execute("/seller/get", "GET", nil)
			if err != nil {
				log.Printf("Failed to fetch store info: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to fetch store details"})
			}

			var storeInfo models.ApiResponseStoreInfo
			jsonData, _ := json.Marshal(resp.Data)
			if err := json.Unmarshal(jsonData, &storeInfo); err != nil {
				log.Printf("Failed to parse store response: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to parse store data"})
			}

			// Insert store details into the database
			insertStoreQuery := `INSERT INTO store (id, name, platform, region, status) VALUES ($1, $2, 'Lazada', $3, $4) RETURNING id`
			err = db.QueryRow(insertStoreQuery, userInfo.SellerID, storeInfo.Name, storeInfo.Location, storeInfo.Status).Scan(&storeID)
			if err != nil {
				log.Printf("Failed to insert store info: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to save store data"})
			}
		} else if err != nil {
			log.Printf("Database query error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Database query error"})
		}

		// Save access token
		insertTokenQuery := `INSERT INTO accessTokens (account_id, store_id, access_token, refresh_token, platform) VALUES ($1, $2, $3, $4, 'Lazada')`
		_, err = db.Exec(insertTokenQuery, userInfo.UserID, storeID, tokenData.AccessToken, tokenData.RefreshToken)
		if err != nil {
			log.Printf("Failed to save access token: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to save access token"})
		}

		// Return response
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":       "Access token successfully stored",
			"user_id":       userInfo.UserID,
			"seller_id":     userInfo.SellerID,
			"is_main":       isMainAccount,
			"store_id":      storeID,
			"access_token":  tokenData.AccessToken,
			"refresh_token": tokenData.RefreshToken,
		})
	}
}

func GetProducts(db *sql.DB, client *sdk.IopClient) echo.HandlerFunc {
	return func(c echo.Context) error {

		resp, _, err := client.Execute("/seller/get", "GET", nil)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch store info"})
		}

		if resp == nil || len(resp.Data) == 0 {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid response from Lazada API"})
		}

		var storeInfo models.ApiResponseStoreInfo
		err = json.Unmarshal(resp.Data, &storeInfo)
		if err != nil {
			log.Println("Error unmarshaling 'data' into ApiResponseStoreInfo:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse store data"})
		}

		return c.JSON(http.StatusOK, storeInfo)
	}
}

func LazadaGetSellerInfo(client *sdk.IopClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, _, err := client.Execute("/seller/get", "GET", nil)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch store info"})
		}

		if resp == nil || len(resp.Data) == 0 {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid response from Lazada API"})
		}

		var storeInfo models.ApiResponseStoreInfo
		err = json.Unmarshal(resp.Data, &storeInfo)
		if err != nil {
			log.Println("Error unmarshaling 'data' into ApiResponseStoreInfo:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse store data"})
		}

		return c.JSON(http.StatusOK, storeInfo)
	}
}
