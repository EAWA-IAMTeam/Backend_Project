package services

import (
	"backend_project/internal/orders/models"
	"backend_project/internal/orders/repositories"
	"fmt"
)

type OrdersService interface {
	GetOrders(createdAfter string, createdBefore string, offset int, limit int, status string, sort_direction string, companyID string, storeID string) ([]models.Order, int, error)
	SaveOrder(order *models.Order, companyID string, storeID string) error
	FetchOrdersByCompanyID(companyID string) ([]models.Order, error)
}

type ordersService struct {
	repo            repositories.OrdersRepository
	itemListService ItemListService
	returnService   ReturnService
	paymentService  PaymentService
}

func NewOrdersService(repo repositories.OrdersRepository, itemListService ItemListService, returnService ReturnService, paymentService PaymentService) OrdersService {
	return &ordersService{repo, itemListService, returnService, paymentService}
}

func (s *ordersService) GetOrders(createdAfter string, createdBefore string, offset int, limit int, status string, sort_direction string, companyID string, storeID string) ([]models.Order, int, error) {
	ordersData, err := s.repo.FetchOrders(createdAfter, createdBefore, offset, limit, status, sort_direction, companyID, storeID)
	if err != nil {
		return nil, 0, err
	}

	if ordersData == nil || len(ordersData.Orders) == 0 {
		return nil, 0, nil
	}

	// Extract order IDs
	var orderIDs []string
	for _, order := range ordersData.Orders {
		orderIDs = append(orderIDs, fmt.Sprintf("%d", order.OrderID))
	}

	// Fetch items for the orders
	items, err := s.itemListService.GetItemList(orderIDs)
	if err != nil {
		return nil, 0, err
	}

	// Map items to their respective orders
	for i, order := range ordersData.Orders {
		for _, item := range items {
			if item.OrderID == order.OrderID {
				ordersData.Orders[i].Items = append(ordersData.Orders[i].Items, item)
			}
		}

		// Fetch and merge return data
		// Only process returns for orders with "returned" status
		hasReturnedStatus := false
		for _, status := range order.Statuses {
			if status == "returned" {
				hasReturnedStatus = true
				break
			}
		}

		if hasReturnedStatus {
			returnData, err := s.returnService.ProcessReturn(fmt.Sprintf("%d", order.OrderID), "1", "1")
			if err == nil {
				ordersData.Orders[i].RefundStatus = returnData
			}
		}
	}

	return ordersData.Orders, ordersData.CountTotal, nil
}

func (s *ordersService) SaveOrder(order *models.Order, companyID string, storeID string) error {
	return s.repo.SaveOrder(order, companyID, storeID)
}

func (s *ordersService) FetchOrdersByCompanyID(companyID string) ([]models.Order, error) {
	return s.repo.FetchOrdersByCompanyID(companyID)
}
