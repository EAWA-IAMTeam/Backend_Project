package models

// Define stores structure

// struct store table in Database
type Store struct {
	ID            int64
	CompanyID     int64
	AccessTokenID int64
	ExpiryTime    string
	Name          string
	Platform      string
	Region        string
	Description   string
	Status        bool
}

// struct account table in Database
type Account struct {
	ID        int64  `json:"id"`
	CompanyID int64  `json:"company_id"`
	Name      string `json:"name"`
	Platform  string `json:"platform"`
	Region    string `json:"region"`
	IsMain    bool   `json:"is_main"`
}

type AccessToken struct {
	ID           int64  `json:"id"`
	AccountID    int64  `json:"account_id"`
	StoreID      int64  `json:"store_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Platform     string `json:"platform"`
}
