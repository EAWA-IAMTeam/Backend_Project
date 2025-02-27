package services

import (
	"backend_project/internal/orders/models"
	"backend_project/internal/orders/repositories"
)

type ItemListService interface {
	GetItemList(orderIDs []string) ([]models.Item, error)
}

type itemListService struct {
	repo repositories.ItemListRepository
}

func NewItemListService(repo repositories.ItemListRepository) ItemListService {
	return &itemListService{repo}
}

func (s *itemListService) GetItemList(orderIDs []string) ([]models.Item, error) {
	orderItems, err := s.repo.FetchItemList(orderIDs)
	if err != nil {
		return nil, err
	}

	items := make([]models.Item, 0)
	for _, orderItem := range orderItems {
		items = append(items, orderItem.OrderItems...)
	}

	return items, nil
}
