package models

import "time"

type StockItem struct {
	ID               int64     `json:"stock_item_id"`
	CompanyID        int       `json:"company_id"`
	StockCode        string    `json:"stock_code"`
	StockControl     bool      `json:"stock_control"`
	RefCost          float64   `json:"ref_cost"`
	RefPrice         float64   `json:"ref_price"`
	ReservedQuantity int       `json:"reserved_quantity"`
	Quantity         int       `json:"quantity"`
	UpdatedAt        time.Time `json:"updated_at"`
	Description      string    `json:"description"`
	Status           bool      `json:"status"`
}

/* GET
[
    {
        "stock_item_id": 1,
        "company_id": 2,
        "stock_code": "STK001",
        "stock_control": true,
        "ref_cost": 90,
        "ref_price": 120.5,
        "reserved_quantity": 20,
        "quantity": 20,
        "updated_at": "0001-01-01T00:00:00Z",
        "description": "Black leather jacket",
        "status": true
    },
    {
        "stock_item_id": 2,
        "company_id": 2,
        "stock_code": "STK002",
        "stock_control": true,
        "ref_cost": 70,
        "ref_price": 89.99,
        "reserved_quantity": 50,
        "quantity": 50,
        "updated_at": "0001-01-01T00:00:00Z",
        "description": "Casual blue denim jeans",
        "status": true
    },
    {
        "stock_item_id": 3,
        "company_id": 2,
        "stock_code": "STK003",
        "stock_control": true,
        "ref_cost": 110,
        "ref_price": 150,
        "reserved_quantity": 10,
        "quantity": 30,
        "updated_at": "0001-01-01T00:00:00Z",
        "description": "Red woolen sweater",
        "status": true
    },
    {
        "stock_item_id": 4,
        "company_id": 2,
        "stock_code": "STK004",
        "stock_control": true,
        "ref_cost": 55,
        "ref_price": 75,
        "reserved_quantity": 5,
        "quantity": 40,
        "updated_at": "0001-01-01T00:00:00Z",
        "description": "Green cotton t-shirt",
        "status": true
    },
    {
        "stock_item_id": 5,
        "company_id": 2,
        "stock_code": "STK005",
        "stock_control": true,
        "ref_cost": 150,
        "ref_price": 200,
        "reserved_quantity": 3,
        "quantity": 15,
        "updated_at": "0001-01-01T00:00:00Z",
        "description": "White down jacket",
        "status": true
    }
]
*/

type StoreProduct struct {
	ID              int64     `json:"id"`
	StoreID         int64     `json:"store_id"`
	StockItemID     int64     `json:"stock_item_id"`
	SKU             string    `json:"sku"`
	Currency        string    `json:"currency"`
	Price           float64   `json:"price"`
	Status          string    `json:"status"`
	ImageURL        []string  `json:"image_url"`
	UpdatedAt       time.Time `json:"updated_at"`
	DiscountedPrice float64   `json:"discounted_price"`
	Weight          string    `json:"weight"`
	Length          string    `json:"length"`
	Width           string    `json:"width"`
	Height          string    `json:"height"`
	Variation1      string    `json:"variation1"`
	Variation2      string    `json:"variation2"`
}

/*POST
[
    {
        "stock_item_id": 149,
        "store_id": 300163725140,
        "price": 29.99,
        "discounted_price": 24.99,
        "sku": "4521449533_MY-25553008340",
        "currency": "MYR",
        "status": "active",
        "image_url": [
                    "https://thumbs.dreamstime.com/b/red-apple-isolated-clipping-path-19130134.jpg",
                    "https://plus.unsplash.com/premium_photo-1661322640130-f6a1e2c36653?fm=jpg&q=60&w=3000&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8MXx8YXBwbGV8ZW58MHx8MHx8fDA%3D"
                ]
		"weight" :"10",
        "length" : "20",
        "width" : "30",
        "height" : "40"
    },
        {
        "stock_item_id": 149,
        "store_id": 3,
        "price": 29.99,
        "discounted_price": 24.99,
        "sku": "4521813030_MY-25554653289",
        "currency": "MYR",
        "status": "active",
        "image_url": [
                    "https://thumbs.dreamstime.com/b/red-apple-isolated-clipping-path-19130134.jpg",
                    "https://plus.unsplash.com/premium_photo-1661322640130-f6a1e2c36653?fm=jpg&q=60&w=3000&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8MXx8YXBwbGV8ZW58MHx8MHx8fDA%3D"
                ]
		"weight" :"10",
        "length" : "20",
        "width" : "30",
        "height" : "40"
    },
        {
        "stock_item_id": 149,
        "store_id": 3,
        "price": 29.99,
        "discounted_price": 24.99,
        "sku": "4521813030_MY-25554653259",
        "currency": "MYR",
        "status": "active",
        "image_url": [
                    "https://thumbs.dreamstime.com/b/red-apple-isolated-clipping-path-19130134.jpg",
                    "https://plus.unsplash.com/premium_photo-1661322640130-f6a1e2c36653?fm=jpg&q=60&w=3000&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8MXx8YXBwbGV8ZW58MHx8MHx8fDA%3D"
                ],
        "weight" :"10",
        "length" : "20",
        "width" : "30",
        "height" : "40"
    }
]
*/

type MergeProduct struct {
	StockItemID   int64          `json:"stock_item_id"`
	RefPrice      float64        `json:"ref_price"`
	RefCost       float64        `json:"ref_cost"`
	Quantity      int            `json:"quantity"`
	StoreProducts []StoreProduct `json:"store_products"`
}

type InsertResult struct {
	Inserted   int      `json:"inserted"`
	Duplicates []string `json:"duplicates"`
}

type Request struct {
	ID              int64     `json:"id"`
	StoreID         int64     `json:"store_id"`
	StockItemID     int64     `json:"stock_item_id"`
	SKU             string    `json:"sku"`
	Currency        string    `json:"currency"`
	Price           float64   `json:"price"`
	Status          string    `json:"status"`
	ImageURL        []string  `json:"image_url"`
	UpdatedAt       time.Time `json:"updated_at"`
	DiscountedPrice float64   `json:"discounted_price"`
	Weight          string    `json:"weight"`
	Width           string    `json:"width"`
	Height          string    `json:"height"`
	Length          string    `json:"length"`
	Variation1      string    `json:"variation1"`
	Variation2      string    `json:"variation2"`
	RequestID       string    `json:"request_id"`
}
