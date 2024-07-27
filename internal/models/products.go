package models

import (
	"encoding/json"
	"log"
	"net/http"

	"First_internet_store/internal/auth"
	"First_internet_store/internal/database"
	"First_internet_store/internal/utils"
)

type PageData struct {
	User
	ProductsData
	UserCookie
}

type Product struct {
	ID         int
	Name       string
	Price      float64
	ImageURL   string
	Quantity   int
	TotalPrice float64 // общая стоимость
}

type ProductsData struct {
	Products []Product
	TotalSum float64 // общая сумма
}

type UserCookie struct {
	UserID   int
	UserName string
}

// главная страница с товарами
func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	// Connect to the database
	db, err := database.Connect()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		log.Println("Error connecting to the database")
		return
	}
	defer db.Close()

	// Выполнение SQL запроса для выборки всех товаров из таблицы "products"
	query := `SELECT name, price, id, image_url FROM products`
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Создание списка для хранения товаров
	var products []Product

	// Считывание данных о товарах из результатов запроса
	for rows.Next() {
		var product Product

		if err := rows.Scan(&product.Name, &product.Price, &product.ID, &product.ImageURL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Добавление продуктов в список
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userName, _ := auth.GetUserName(r)

	data := PageData{}.newPageDataProd(products, userName)

	utils.RenderTemplate(w, data,
		"web/html/main.html",
		"web/html/navigation.html",
	)
}

// создание данных для страницы продуктов и куки пользователя
func (pd PageData) newPageDataProd(products []Product, userName string) PageData {
	return PageData{
		ProductsData: ProductsData{
			Products: products,
		},
		UserCookie: UserCookie{UserName: userName},
	}
}

// Обработчик для добавления товара в корзину
func AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	db, err := database.Connect()
	if err != nil {
		log.Println("Error connecting to the database:", err)
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var requestData struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		log.Println("Invalid request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userName, err := auth.GetUserName(r)
	if err != nil {
		log.Println("User not authenticated:", err)
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var userID int
	err = db.QueryRow("SELECT user_id FROM users WHERE username = $1", userName).Scan(&userID)
	if err != nil {
		log.Println("User not found:", err)
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(`
        INSERT INTO carts (user_id, product_id, quantity) 
        	VALUES ($1, $2, $3) 
       		ON CONFLICT (user_id, product_id) 
        	DO UPDATE SET quantity = carts.quantity + EXCLUDED.quantity`,
		userID, requestData.ProductID, requestData.Quantity,
	)
	if err != nil {
		log.Println("Error adding product to cart:", err)
		http.Error(w, "Error adding product to cart", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
