package services

import (
	"backend_project/internal/payment/models"
	"backend_project/internal/payment/repositories"
)

type ReturnService interface {
	ProcessReturn(trade_order_id string, page_no string, page_size string) ([]models.ReturnRefund, error)
}

type returnService struct {
	repo repositories.ReturnRepository
}

func NewReturnService(repo repositories.ReturnRepository) ReturnService {
	return &returnService{repo}
}

func (s *returnService) ProcessReturn(trade_order_id string, page_no string, page_size string) ([]models.ReturnRefund, error) {
	returnData, err := s.repo.ProcessReturn(trade_order_id, page_no, page_size)
	if err != nil {
		return nil, err
	}

	processedReturnData := make([]models.ReturnRefund, 0)
	for _, item := range returnData.Items {
		processedReturnData = append(processedReturnData, item.ReverseOrderLines...)
	}

	return processedReturnData, nil
}