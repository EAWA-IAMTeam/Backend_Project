package services

import (
	"backend_project/internal/payment/models"
	"backend_project/internal/payment/repositories"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type PaymentService interface {
	// GetOrder(order_id int) ([]models.LazadaOrder, error)
	GetTransactions(startTime string, endTime string, orderID string, offset int, limit int) ([]models.LazadaTransaction, error)
	GetPayouts(createdAfter string) ([]models.LazadaPayout, error)
	CalculateTotalAmountByOrder(resp *models.SQLData, transactions []models.LazadaTransaction)
	AssignPaymentDataToPayout(resp *models.LazadaAPIResponse, transactions []models.LazadaTransaction)
	ExtractDateFromStatement(statement string) string
	ExtractDateFromStatementNumber(statement string) string
}

type paymentService struct {
	paymentsRepository repositories.PaymentRepository
}

// Constructor
func NewPaymentService(paymentsRepository repositories.PaymentRepository) PaymentService {
	return &paymentService{paymentsRepository: paymentsRepository}
}

// FUNCTION PART--------------------------------------------------------------

// GetOrders
// func (p *paymentService) GetOrders(createdAfter string, offset int, limit int, status string) ([]models.LazadaOrder, error) {
// 	orderData, err := p.paymentRepository.FetchOrders(createdAfter, offset, limit, status)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if orderData == nil || len(orderData.Order) == 0 {
// 		log.Printf("No orders found after %s with status %s", createdAfter, status)
// 		return []models.LazadaOrder{}, nil
// 	}
// 	log.Printf("Parsed %d orders (Created After: %s)\n", len(orderData.Order), createdAfter)
// 	return orderData.Order, nil
// }

// GetTransactions
func (p *paymentService) GetTransactions(startTime string, endTime string, orderID string, offset int, limit int) ([]models.LazadaTransaction, error) {
	// Fetch Transactions
	transactions, err := p.paymentsRepository.FetchTransactions(startTime, endTime, orderID, offset, limit)
	if err != nil {
		log.Printf("Error fetching transactions for order %s: %v", orderID, err)
		return nil, err
	}

	if transactions == nil || len(transactions.Transaction) == 0 {
		log.Println("No transactions found.")
		return []models.LazadaTransaction{}, nil
	}

	log.Printf("Returning %d transactions\n", len(transactions.Transaction))
	return transactions.Transaction, nil
}

// GetPayout
func (p *paymentService) GetPayouts(createdAfter string) ([]models.LazadaPayout, error) {
	// Fetch Payout
	payouts, err := p.paymentsRepository.FetchPayouts(createdAfter)
	if err != nil {
		log.Print("Error fetching payout ", err)
		return nil, err
	}
	if payouts == nil || len(payouts.Payout) == 0 {
		log.Println("No payout found.")
		return []models.LazadaPayout{}, nil
	}
	log.Printf("Returning %d payouts\n", len(payouts.Payout))
	return payouts.Payout, nil
}

// Assign Payment Data to Orders
// func (p *paymentService) AssignPaymentDataToOrder(resp *models.LazadaAPIResponse, transactions []models.LazadaTransaction) {
// 	// Initialize the map for lookup PaymentData by OrderID
// 	paymentIDMap := make(map[int64][]models.LazadaTransaction)
// 	// Convert OrderID(string) to integer and put into map
// 	for _, payment := range transactions {
// 		orderIDInInt, err := strconv.ParseInt(payment.OrderNo, 10, 64)
// 		if err != nil {
// 			log.Printf("Error converting orderID '%s' to integer: %v", payment.OrderNo, err)
// 			continue
// 		}
// 		paymentIDMap[orderIDInInt] = append(paymentIDMap[orderIDInInt], payment)
// 	}
// 	// Assign transaction data to orders based on OrderID
// 	for i := range resp.Order {
// 		if data, exists := paymentIDMap[resp.Order[i].OrderNumber]; exists {
// 			resp.Order[i].Transactions = append(resp.Order[i].Transactions, data...)
// 			// Calculate ActualPayment (sum of all transaction amounts)
// 			var totalPayment float64
// 			for _, transaction := range data {
// 				amount, err := strconv.ParseFloat(transaction.Amount, 64)
// 				if err != nil {
// 					log.Printf("Error converting amount for Order %d: %v", resp.Order[i].OrderNumber, err)
// 					continue
// 				}
// 				totalPayment += amount
// 			}
// 			resp.Order[i].ActualPayment = totalPayment
// 		}
// 	}
// }

func (p *paymentService) CalculateTotalAmountByOrder(resp *models.SQLData, transactions []models.LazadaTransaction) {
	// Initialize total payment variable
	var totalPayment float64

	// Sum up the transaction amounts for the given OrderID
	for _, payment := range transactions {
		orderIDInInt, err := strconv.ParseInt(payment.OrderNo, 10, 64)
		if err != nil {
			log.Printf("Error converting orderID '%s' to integer: %v", payment.OrderNo, err)
			continue
		}

		// Check if transaction matches the current order
		if orderIDInInt == resp.OrderID {
			amount, err := strconv.ParseFloat(payment.Amount, 64)
			if err != nil {
				log.Printf("Error converting amount for Order %d: %v", orderIDInInt, err)
				continue
			}

			totalPayment += amount
		}
	}

	// Assign the total actual payment to the order
	resp.TotalReleasedAmount = totalPayment
}


func (p *paymentService) AssignPaymentDataToPayout(resp *models.LazadaAPIResponse, transactions []models.LazadaTransaction) {
	/*
		Formulae to connect both transaction and payout
		Statement at transaction: 21 Oct 2024 - 21 Oct 2024
		StatementNmber at payout: MY4NA1T7CK-2024-1021
	*/
	// Initialize map for lookup PaymentData by statement num
	statementNumberMap := make(map[string][]models.LazadaTransaction)
	// Extract statement to date format
	for _, statement := range transactions {
		statementNumberUnformatted := p.ExtractDateFromStatement(statement.Statement)
		if statementNumberUnformatted == "" {
			log.Print("Error extracting date from statement")
			continue
		}
		statementNumberMap[statementNumberUnformatted] = append(statementNumberMap[statementNumberUnformatted], statement)
	}
	// Assign transaction data to payout based on statement formulae
	for i := range resp.Payout {
		if data, exists := statementNumberMap[p.ExtractDateFromStatementNumber(resp.Payout[i].StatementNumber)]; exists {
			resp.Payout[i].Transactions = append(resp.Payout[i].Transactions, data...)
		}
	}
}

// Extract Date Functions
func (p *paymentService) ExtractDateFromStatement(statement string) string {
	//Split statement of transaction
	transactionStatement := strings.Split(statement, " - ")
	if len(transactionStatement) < 1 {
		log.Print("Invalid statement date format", statement)
		return ""
	}

	dateStr := transactionStatement[0]
	date, err := time.Parse("2 Jan 2006", dateStr)
	if err != nil {
		log.Print("Error parsing date:", err)
		return ""
	}

	return date.Format("2006-0102")
}

func (p *paymentService) ExtractDateFromStatementNumber(statement string) string {
	//Split statement of payout using regex
	statementRegex := regexp.MustCompile(`\d{4}-\d{4}`)
	match := statementRegex.FindString(statement)
	if match == "" {
		log.Print("Invalid statement number format", statement)
		return ""
	}
	return match
}
