package models

type PaginatedRequest struct {
	CompanyID  int64  `json:"company_id"`
	RequestID  string `json:"request_id"`
	Pagination struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
	} `json:"pagination"`
}
