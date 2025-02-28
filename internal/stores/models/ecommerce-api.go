package models

// From ApiResponseAccessToken transform into LinkStore format
type LinkStore struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	Country          string `json:"country"`
	UserID           string `json:"user_id"`
	SellerID         string `json:"seller_id"`
	Account          string `json:"account"`
	ShortCode        string `json:"short_code"`
}

/*
	below struct is Lazada API response for 'auth/token/create' endpoint
*/
type ApiResponseAccessToken struct {
	AccessToken      string     `json:"access_token"`
	Country          string     `json:"country"`
	RefreshToken     string     `json:"refresh_token"`
	AccountPlatform  string     `json:"account_platform"`
	RefreshExpiresIn int        `json:"refresh_expires_in"`
	UserInfo         []UserInfo `json:"country_user_info"`
	ExpiresIn        int        `json:"expires_in"`
	Account          string     `json:"account"`
	Code             string     `json:"code"`
	RequestID        string     `json:"request_id"`
}

type UserInfo struct {
	Country   string `json:"country"`
	UserID    string `json:"user_id"`
	SellerID  string `json:"seller_id"`
	ShortCode string `json:"short_code"`
}

/*
	below struct is Lazada API response for 'seller/get' endpoint
*/
type ApiResponseStoreInfo struct {
	Name                string `json:"name"`
	Verified            bool   `json:"verified"`
	Location            string `json:"location"`
	MarketPlaceEaseMode bool   `json:"marketplaceEaseMode"`
	SellerID            int64  `json:"seller_id"`
	Email               string `json:"email"`
	ShortCode           string `json:"short_code"`
	CB                  bool   `json:"cb"`     // Cross Border seller
	Status              string `json:"status"` // ACTIVE INACTIVE DELETED
}

type LazadaStoreResponse struct {
	Code    string               `json:"code"`
	Message string               `json:"message"`
	Data    ApiResponseStoreInfo `json:"data"`
}
