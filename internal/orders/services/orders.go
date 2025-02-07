package services

import (
	"backend_project/internal/orders/models"
	"backend_project/internal/orders/repositories"
	"log"
)

type OrdersService interface {
	GetOrders(createdAfter string) ([]models.Order, error)
}

type ordersService struct {
	repo repositories.OrdersRepository
}

func NewOrdersService(repo repositories.OrdersRepository) OrdersService {
	return &ordersService{repo}
}

func (s *ordersService) GetOrders(createdAfter string) ([]models.Order, error) {
	ordersData, err := s.repo.FetchOrders(createdAfter)
	if err != nil {
		return nil, err
	}

	if ordersData == nil || len(ordersData.Orders) == 0 {
		log.Println("No orders found in response")
		return nil, nil
	}

	log.Printf("Parsed orders count: %d\n", len(ordersData.Orders))
	return ordersData.Orders, nil
}
