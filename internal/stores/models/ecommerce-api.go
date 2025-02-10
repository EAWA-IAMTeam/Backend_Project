package models

//Lazada struct
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

type ApiResponseStoreInfo struct {
	Name                string `json:"name"`
	Verified            bool   `json:"verified"`
	Location            string `json:"location"`
	MarketPlaceEaseMode bool   `json:"marketplaceEaseMode"`
	SellerID            int    `json:"seller_id"`
	Email               string `json:"email"`
	ShortCode           string `json:"short_code"`
	CB                  bool   `json:"cb"`     // Cross Border seller
	Status              string `json:"status"` // ACTIVE INACTIVE DELETED
}
