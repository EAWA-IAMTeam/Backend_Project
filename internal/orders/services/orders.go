package services

import (
	"backend_project/internal/orders/models"
	"backend_project/internal/orders/repositories"
	"fmt"
	"log"
)

type OrdersService interface {
	GetOrders(createdAfter string, offset int, limit int, status string) ([]models.Order, error)
	SaveOrder(order *models.Order, companyID string) error
}

type ordersService struct {
	repo            repositories.OrdersRepository
	itemListService ItemListService
}

func NewOrdersService(repo repositories.OrdersRepository, itemListService ItemListService) OrdersService {
	return &ordersService{repo, itemListService}
}

func (s *ordersService) GetOrders(createdAfter string, offset int, limit int, status string) ([]models.Order, error) {
	ordersData, err := s.repo.FetchOrders(createdAfter, offset, limit, status)
	if err != nil {
		return nil, err
	}

	if ordersData == nil || len(ordersData.Orders) == 0 {
		log.Println("No orders found in response")
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
	}

	log.Printf("Parsed orders count: %d\n", len(ordersData.Orders))
	return ordersData.Orders, nil
}

func (s *ordersService) SaveOrder(order *models.Order, companyID string) error {
	return s.repo.SaveOrder(order, companyID)
}
