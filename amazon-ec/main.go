package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/glebarez/go-sqlite"
)

type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

type Order struct {
	ID        int    `json:"id"`
	ProductID int    `json:"product_id"`
	Quantity  int    `json:"quantity"`
	CreatedAt string `json:"created_at"`
}

type OrderDetail struct {
	ID          int    `json:"id"`
	ProductID   int    `json:"product_id"`
	ProductName string `json:"product_name"`
	Price       int    `json:"price"`
	Quantity    int    `json:"quantity"`
	TotalPrice  int    `json:"total_price"`
	CreatedAt   string `json:"created_at"`
}

type PaymentRequest struct {
	OrderID int    `json:"order_id"`
	Amount  int    `json:"amount"`
	Method  string `json:"method"`
}

type Payment struct {
	ID        int    `json:"id"`
	OrderID   int    `json:"order_id"`
	Amount    int    `json:"amount"`
	Method    string `json:"method"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite", "./ec.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable()

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/products", getProductsHandler)
	http.HandleFunc("/api/product", getProductHandler)
	http.HandleFunc("/api/orders", ordersHandler)
	http.HandleFunc("/api/payments", paymentsHandler)

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createTable() {
	createProductsTable := `
	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		price INTEGER NOT NULL,
		description TEXT,
		image_url TEXT
	);`

	createOrdersTable := `
	CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		product_id INTEGER NOT NULL,
		quantity INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (product_id) REFERENCES products(id)
	);`

	createPaymentsTable := `
	CREATE TABLE IF NOT EXISTS payments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id INTEGER NOT NULL,
		amount INTEGER NOT NULL,
		method TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(createProductsTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createOrdersTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createPaymentsTable)
	if err != nil {
		log.Fatal(err)
	}

	insertSampleData()
}

func insertSampleData() {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count == 0 {
		sampleProducts := []Product{
			{Name: "ノートパソコン", Price: 80000, Description: "高性能で軽量なノートパソコンです。", ImageURL: "https://via.placeholder.com/150"},
			{Name: "スマートフォン", Price: 50000, Description: "最新機能が詰まったスマートフォンです。", ImageURL: "https://via.placeholder.com/150"},
			{Name: "ワイヤレスイヤホン", Price: 12000, Description: "高音質でノイズキャンセリング付きです。", ImageURL: "https://via.placeholder.com/150"},
		}

		for _, p := range sampleProducts {
			_, err := db.Exec("INSERT INTO products (name, price, description, image_url) VALUES (?, ?, ?, ?)", p.Name, p.Price, p.Description, p.ImageURL)
			if err != nil {
				log.Fatal(err)
			}
		}
		log.Println("Sample products inserted successfully.")
	}
}

func getProductsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := db.Query("SELECT id, name, price, description, image_url FROM products")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Description, &p.ImageURL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}

	if len(products) == 0 {
		products = []Product{}
	}

	json.NewEncoder(w).Encode(products)
}

func getProductHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var p Product
	err = db.QueryRow("SELECT id, name, price, description, image_url FROM products WHERE id = ?", id).Scan(&p.ID, &p.Name, &p.Price, &p.Description, &p.ImageURL)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(p)
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		query := `
		SELECT 
			o.id, 
			o.product_id, 
			p.name, 
			p.price, 
			o.quantity, 
			(p.price * o.quantity) AS total_price, 
			o.created_at
		FROM orders o
		JOIN products p ON o.product_id = p.id
		ORDER BY o.id DESC
		`

		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var orders []OrderDetail
		for rows.Next() {
			var o OrderDetail
			if err := rows.Scan(&o.ID, &o.ProductID, &o.ProductName, &o.Price, &o.Quantity, &o.TotalPrice, &o.CreatedAt); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			orders = append(orders, o)
		}

		if len(orders) == 0 {
			orders = []OrderDetail{}
		}

		json.NewEncoder(w).Encode(orders)
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			ProductID int `json:"product_id"`
			Quantity  int `json:"quantity"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := db.Exec("INSERT INTO orders (product_id, quantity) VALUES (?, ?)", req.ProductID, req.Quantity)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id, err := res.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "Order created successfully",
			"order_id": id,
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func paymentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.OrderID <= 0 || req.Amount <= 0 {
		http.Error(w, "Invalid order_id or amount", http.StatusBadRequest)
		return
	}

	if req.Method != "credit" && req.Method != "bank" && req.Method != "cod" {
		http.Error(w, "Invalid payment method", http.StatusBadRequest)
		return
	}

	status := "paid"

	res, err := db.Exec(
		"INSERT INTO payments (order_id, amount, method, status) VALUES (?, ?, ?, ?)",
		req.OrderID, req.Amount, req.Method, status,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Payment processed successfully",
		"payment_id": id,
		"order_id":   req.OrderID,
		"amount":     req.Amount,
		"method":     req.Method,
		"status":     status,
	})
}
