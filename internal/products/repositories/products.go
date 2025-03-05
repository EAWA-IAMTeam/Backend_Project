package repositories

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/lib/pq"

	"backend_project/internal/products/models"
)

type ProductRepository struct {
	DB *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{DB: db}
}

// GetStockItemsByCompany fetches stock items by company ID
func (pr *ProductRepository) GetStockItemsByCompany(companyID int64, page, limit int) ([]*models.StockItem, error) {
	//Calculate offset
	offset := (page - 1) * limit

	query := `
	SELECT id, company_id, stock_code, stock_control, ref_price, ref_cost, quantity, reserved_quantity,
		   description, status 
	FROM stockitem 
	WHERE company_id = $1
	LIMIT $2 OFFSET $3`
	// query := `
	//     SELECT id, company_id, stock_code, stock_control, ref_price, ref_cost, weight,
	//            height, width, length, variation1, variation2, quantity, reserved_quantity,
	//            platform, description, status
	//     FROM stockitem
	//     WHERE company_id = $1`

	rows, err := pr.DB.Query(query, companyID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stockItems []*models.StockItem
	for rows.Next() {
		var item models.StockItem
		if err := rows.Scan(
			&item.ID, &item.CompanyID, &item.StockCode, &item.StockControl, &item.RefPrice,
			&item.RefCost, &item.Quantity, &item.ReservedQuantity, &item.Description, &item.Status,
		); err != nil {
			return nil, err
		}
		stockItems = append(stockItems, &item)
	}

	return stockItems, nil
}

func (pr *ProductRepository) InsertStockItemsByCompany(companyID int64, stockItems []models.StockItem) error {
	tx, err := pr.DB.Begin()
	if err != nil {
		fmt.Println("Error starting transaction:", err)
		return err
	}

	query := `
		INSERT INTO stockitem 
		(company_id, ref_price, ref_cost, quantity, reserved_quantity, stock_code, stock_control, 
		description, status, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (company_id, stock_code) 
		DO UPDATE SET 
			ref_price = EXCLUDED.ref_price,
			ref_cost = EXCLUDED.ref_cost,
			quantity = EXCLUDED.quantity,
			reserved_quantity = EXCLUDED.reserved_quantity,
			stock_control = EXCLUDED.stock_control,
			description = EXCLUDED.description,
			status = EXCLUDED.status,
			updated_at = NOW()
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		fmt.Println("Error preparing statement:", err)
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	stockCodes := make(map[string]bool)

	for _, item := range stockItems {
		stockCodes[item.StockCode] = true
		_, err := stmt.Exec(
			companyID, item.RefPrice, item.RefCost, item.Quantity, item.ReservedQuantity,
			item.StockCode, item.StockControl, item.Description, item.Status,
		)
		if err != nil {
			fmt.Println("Error executing insert query:", err)
			tx.Rollback()
			return err
		}
	}

	// Remove stock items not in the request
	deleteQuery := `
		DELETE FROM stockitem 
		WHERE company_id = $1 AND stock_code NOT IN (SELECT unnest($2::text[]))
	`

	stockCodeList := make([]string, 0, len(stockCodes))
	for code := range stockCodes {
		stockCodeList = append(stockCodeList, code)
	}

	fmt.Println("Deleting stock items not in request for company:", companyID)
	_, err = tx.Exec(deleteQuery, companyID, pq.Array(stockCodeList))
	if err != nil {
		fmt.Println("Error executing delete query:", err)
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		fmt.Println("Error committing transaction:", err)
		return err
	}

	fmt.Println("Stock items inserted/updated successfully for company:", companyID)
	return nil
}

// GetProductsByCompany fetches products by company ID
func (pr *ProductRepository) GetProductsByCompany(companyID int64, page, limit int) ([]*models.MergeProduct, error) {
	// Step 1: Fetch all store IDs for the given company
	storeQuery := `SELECT store_id FROM store WHERE company_id = $1`
	storeRows, err := pr.DB.Query(storeQuery, companyID)
	if err != nil {
		return nil, err
	}
	defer storeRows.Close()

	var storeIDs []int
	for storeRows.Next() {
		var storeID int
		if err := storeRows.Scan(&storeID); err != nil {
			return nil, err
		}
		storeIDs = append(storeIDs, storeID)
	}

	log.Println("Hello", storeIDs)
	// If no stores found, return empty result
	if len(storeIDs) == 0 {
		return []*models.MergeProduct{}, nil
	}

	//Calculate offset
	offset := (page - 1) * limit

	// Step 2: Fetch all store products based on store IDs
	// query := `
	//     SELECT si.id, si.ref_price, si.ref_cost, si.quantity,
	//            sp.id, sp.price, sp.discounted_price, sp.sku, sp.currency, sp.status, sp.store_id, sp.media_url
	//     FROM storeproduct sp
	//     JOIN stockitem si ON sp.stock_item_id = si.id
	//     WHERE sp.store_id = ANY($1)`
	query := `
        SELECT si.id, si.ref_price, si.ref_cost, si.quantity,
               sp.id, sp.price, sp.discounted_price, sp.sku, sp.currency, sp.status, sp.store_id, sp.media_url
        FROM storeproduct sp
        JOIN stockitem si ON sp.stock_item_id = si.id
        WHERE sp.store_id = ANY($1)
		LIMIT $2 OFFSET $3`

	rows, err := pr.DB.Query(query, pq.Array(storeIDs), limit, offset) // Using pq.Array for PostgreSQL IN clause
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stockItemsMap := make(map[int64]*models.MergeProduct)
	for rows.Next() {
		var stockItem models.MergeProduct
		var storeProduct models.StoreProduct
		var imageURLs string

		err := rows.Scan(
			&stockItem.StockItemID, &stockItem.RefPrice, &stockItem.RefCost, &stockItem.Quantity,
			&storeProduct.ID, &storeProduct.Price, &storeProduct.DiscountedPrice, &storeProduct.SKU,
			&storeProduct.Currency, &storeProduct.Status, &storeProduct.StoreID, &imageURLs,
		)
		if err != nil {
			return nil, err
		}

		storeProduct.StockItemID = stockItem.StockItemID

		// log.Println(storeProduct)
		// Unmarshal JSON string of image URLs into []string
		if imageURLs != "" {
			err = json.Unmarshal([]byte(imageURLs), &storeProduct.ImageURL)
			if err != nil {
				storeProduct.ImageURL = []string{} // Ensure it's an empty slice on error
			}
		} else {
			storeProduct.ImageURL = []string{} // Set empty slice if imageURLs is empty
		}

		if _, exists := stockItemsMap[stockItem.StockItemID]; !exists {
			stockItemsMap[stockItem.StockItemID] = &stockItem
		}

		stockItemsMap[stockItem.StockItemID].StoreProducts = append(stockItemsMap[stockItem.StockItemID].StoreProducts, storeProduct)
	}

	// Convert map to slice
	var result []*models.MergeProduct
	for _, item := range stockItemsMap {
		result = append(result, item)
	}

	// log.Println(result)
	return result, nil
}

// InsertProductBatch inserts multiple products into the database
func (pr *ProductRepository) InsertProductBatch(products []*models.Request) (*models.InsertResult, error) {
	result := &models.InsertResult{
		Inserted:   0,
		Duplicates: make([]string, 0),
	}

	tx, err := pr.DB.Begin()
	if err != nil {
		return result, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
        INSERT INTO storeproduct (store_id, stock_item_id, sku, currency, price, status, media_url, updated_at, discounted_price, weight, width, height, length  )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
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

		// Convert []string to JSON
		mediaURLJSON, err := json.Marshal(product.ImageURL)
		if err != nil {
			return result, err
		}

		res, err := stmt.Exec(
			product.StoreID,
			product.StockItemID,
			product.SKU,
			product.Currency,
			product.Price,
			product.Status,
			string(mediaURLJSON),
			product.UpdatedAt,
			product.DiscountedPrice,
			product.Weight,
			product.Width,
			product.Height,
			product.Length,
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

	if err := tx.Commit(); err != nil {
		return result, err
	}

	return result, nil
}

// GetStoreByCompany fetches Store IDs for a given company from the database with platform as key
func (r *ProductRepository) GetStoreByCompany(companyID int64) (map[string][]int64, error) {
	query := "SELECT store_id, platform FROM store WHERE company_id = $1"
	rows, err := r.DB.Query(query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()
	log.Print("CompanyID:", companyID)

	// Map to store platform-wise store IDs
	storeMap := make(map[string][]int64)

	for rows.Next() {
		var id int64
		var platform string
		if err := rows.Scan(&id, &platform); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Append store ID to the appropriate platform list
		storeMap[platform] = append(storeMap[platform], id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return storeMap, nil
}

// GetStoreSkus fetches SKUs for a given store from the database
func (r *ProductRepository) GetStoreSkus(storeID int64) (map[string]bool, error) {
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

func (r *ProductRepository) DeleteMappedProductsBySKU(storeID int64, sku string) (int64, error) {
	query := "DELETE FROM storeproduct WHERE store_id = $1 AND sku = $2"
	result, err := r.DB.Exec(query, storeID, sku)
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

func (r *ProductRepository) DeleteMappedProductsBySKUs(storeID int64, skus []string) ([]string, []string, error) {
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
