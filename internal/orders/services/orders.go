package services

import (
	"backend_project/internal/orders/models"
	"backend_project/internal/orders/repositories"
	"fmt"
)

type OrdersService interface {
	GetOrders(createdAfter string, offset int, limit int, status string) ([]models.Order, error)
	SaveOrder(order *models.Order, companyID string) error
}

type ordersService struct {
	repo            repositories.OrdersRepository
	itemListService ItemListService
	returnService   ReturnService
}

func NewOrdersService(repo repositories.OrdersRepository, itemListService ItemListService, returnService ReturnService) OrdersService {
	return &ordersService{repo, itemListService, returnService}
}

func (s *ordersService) GetOrders(createdAfter string, offset int, limit int, status string) ([]models.Order, error) {
	ordersData, err := s.repo.FetchOrders(createdAfter, offset, limit, status)
	if err != nil {
		return nil, err
	}

	if ordersData == nil || len(ordersData.Orders) == 0 {
		return nil, nil
	}

	// Extract order IDs
	var orderIDs []string
	for _, order := range ordersData.Orders {
		orderIDs = append(orderIDs, fmt.Sprintf("%d", order.OrderID))
	}

	// Fetch items for the orders
	items, err := s.itemListService.GetItemList(orderIDs)
	if err != nil {
		return nil, err
	}

	// Map items to their respective orders
	for i, order := range ordersData.Orders {
		for _, item := range items {
			if item.OrderID == order.OrderID {
				ordersData.Orders[i].Items = append(ordersData.Orders[i].Items, item)
			}
		}

		// Fetch and merge return data
		returnData, err := s.returnService.ProcessReturn(fmt.Sprintf("%d", order.OrderID), "1", "1") // Adjust page_size if needed
		if err == nil {
			ordersData.Orders[i].RefundStatus = returnData // Ensure this matches the type in the Order struct
		}
	}

	return ordersData.Orders, nil
}

func (s *ordersService) SaveOrder(order *models.Order, companyID string) error {
	return s.repo.SaveOrder(order, companyID)
}
