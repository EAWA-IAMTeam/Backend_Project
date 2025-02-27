package models

type PaginatedRequest struct {
	CompanyID      int64  `json:"company_id"`
	RequestID      string `json:"request_id"`
	Status         string `json:"status"`
	Created_after  string `json:"created_after"`
	Stop_after     string `json:"stop_after"`
	Sort_direction string `json:"sort_direction"`
	Pagination     struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
	} `json:"pagination"`
}
