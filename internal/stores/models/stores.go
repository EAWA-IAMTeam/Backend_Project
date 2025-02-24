package models

import "time"

// Define stores structure
// struct store table in Database
type Store struct {
	ID            string    `json:"id"`
	CompanyID     int64     `json:"company_id"`
	AccessTokenID int8      `json:"access_token_id"`
	AuthTime      time.Time `json:"authorize_time"`
	ExpiryTime    time.Time `json:"expiry_time"`
	Name          string    `json:"name"`
	Platform      string    `json:"platform"`
	Region        string    `json:"region"`
	Descriptions  string    `json:"descriptions"`
	Status        bool      `json:"status"`
}

// struct account table in Database
type Account struct {
	ID        string `json:"id"`
	CompanyID int64  `json:"company_id"`
	Name      string `json:"name"`
	Platform  string `json:"platform"`
	Region    string `json:"region"`
	IsMain    bool   `json:"is_main"`
}

type AccessToken struct {
	ID           int8   `json:"id"`
	AccountID    string `json:"account_id"`
	StoreID      string `json:"store_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Platform     string `json:"platform"`
}
