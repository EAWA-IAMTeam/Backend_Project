package repositories

import (
	"backend_project/internal/products/models"
	"database/sql"
	"errors"
	"fmt"
)

type ProductRepository struct {
	DB *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{DB: db}
}

// GetStockItemsByCompany fetches stock items by company ID
func (pr *ProductRepository) GetStockItemsByCompany(companyID int) ([]*models.StockItem, error) {
	query := `
        SELECT id, company_id, stock_code, stock_control, ref_price, ref_cost, weight, 
               height, width, length, variation1, variation2, quantity, reserved_quantity,
               platform, description, status 
        FROM stockitem 
        WHERE company_id = $1`

	rows, err := pr.DB.Query(query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stockItems []*models.StockItem
	for rows.Next() {
		var item models.StockItem
		if err := rows.Scan(
			&item.ID, &item.CompanyID, &item.StockCode, &item.StockControl, &item.RefPrice,
			&item.RefCost, &item.Weight, &item.Height, &item.Width,
			&item.Length, &item.Variation1, &item.Variation2, &item.Quantity, &item.ReservedQuantity,
			&item.Platform, &item.Description, &item.Status,
		); err != nil {
			return nil, err
		}
		stockItems = append(stockItems, &item)
	}

	return stockItems, nil
}

// GetProductsByStore fetches products by store ID
func (pr *ProductRepository) GetProductsByStore(storeID int) ([]*models.MergeProduct, error) {
	query := `
        SELECT si.id, si.ref_price, si.ref_cost, si.quantity,
               sp.id, sp.price, sp.discounted_price, sp.sku, sp.currency, sp.status
        FROM storeproduct sp
        JOIN stockitem si ON sp.stock_item_id = si.id
        WHERE sp.store_id = $1`

	rows, err := pr.DB.Query(query, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stockItemsMap := make(map[int64]*models.MergeProduct)
	for rows.Next() {
		var stockItem models.MergeProduct
		var storeProduct models.StoreProduct

		err := rows.Scan(
			&stockItem.StockItemID, &stockItem.RefPrice, &stockItem.RefCost, &stockItem.Quantity,
			&storeProduct.ID, &storeProduct.Price, &storeProduct.DiscountedPrice, &storeProduct.SKU, &storeProduct.Currency, &storeProduct.Status,
		)
		if err != nil {
			return nil, err
		}

		if _, exists := stockItemsMap[stockItem.StockItemID]; !exists {
			stockItemsMap[stockItem.StockItemID] = &stockItem
		}

		stockItemsMap[stockItem.StockItemID].StoreProducts = append(stockItemsMap[stockItem.StockItemID].StoreProducts, storeProduct)
	}

	var result []*models.MergeProduct
	for _, item := range stockItemsMap {
		result = append(result, item)
	}

	return result, nil
}

// InsertProductBatch inserts multiple products into the database
func (pr *ProductRepository) InsertProductBatch(storeID int64, products []models.StoreProduct) (*models.InsertResult, error) {
	result := &models.InsertResult{
		Inserted:   0,
		Duplicates: make([]string, 0),
	}

	// Start a transaction
	tx, err := pr.DB.Begin()
	if err != nil {
		return result, err
	}
	defer tx.Rollback()

	// Prepare the insert statement
	stmt, err := tx.Prepare(`
        INSERT INTO storeproduct (store_id, stock_item_id, price, discounted_price, sku, currency, status)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `)
	if err != nil {
		return result, err
	}
	defer stmt.Close()

	// Check for duplicates first
	for _, product := range products {
		var exists bool
		err := tx.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM storeproduct WHERE stock_item_id = $1 AND sku = $2)",
			product.StockItemID, product.SKU,
		).Scan(&exists)
		if err != nil {
			return result, err
		}
		if exists {
			result.Duplicates = append(result.Duplicates, product.SKU)
			continue
		}

		// Insert the product
		res, err := stmt.Exec(
			storeID,
			product.StockItemID,
			product.Price,
			product.DiscountedPrice,
			product.SKU,
			product.Currency,
			product.Status,
		)
		if err != nil {
			return result, err
		}

		affected, err := res.RowsAffected()
		if err != nil {
			return result, err
		}
		result.Inserted += int(affected)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return result, err
	}

	return result, nil
}

// GetStoreSkus fetches SKUs for a given store from the database
func (r *ProductRepository) GetStoreSkus(storeID string) (map[string]bool, error) {
	query := "SELECT sku FROM storeproduct WHERE store_id = $1"
	rows, err := r.DB.Query(query, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %v", err)
	}
	defer rows.Close()

	skus := make(map[string]bool)
	for rows.Next() {
		var sku string
		if err := rows.Scan(&sku); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		skus[sku] = true
	}

	return skus, rows.Err()
}

func (r *ProductRepository) DeleteMappedProductsBySKU(storeID, sku string) (int64, error) {
	query := "DELETE FROM storeproduct WHERE store_id = $1 AND sku = $2"
	result, err := r.DB.Exec(query, storeID, sku)
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected() // Get number of affected rows
	return rowsAffected, nil
}

func (r *ProductRepository) DeleteMappedProductsBySKUs(storeID string, skus []string) ([]string, []string, error) {
	if len(skus) == 0 {
		return nil, nil, errors.New("no SKUs provided")
	}

	var deletedSKUs []string
	var failedSKUs []string

	// Iterate over each SKU to delete individually
	for _, sku := range skus {
		query := "DELETE FROM storeproduct WHERE store_id = $1 AND sku = $2"
		result, err := r.DB.Exec(query, storeID, sku)
		if err != nil {
			failedSKUs = append(failedSKUs, sku)
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			deletedSKUs = append(deletedSKUs, sku)
		} else {
			failedSKUs = append(failedSKUs, sku)
		}
	}

	return deletedSKUs, failedSKUs, nil
}
