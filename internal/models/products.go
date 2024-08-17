package models

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/KrepkiyOrex/inter_store/internal/auth"
	"github.com/KrepkiyOrex/inter_store/internal/database"
	"github.com/KrepkiyOrex/inter_store/internal/utils"

	"github.com/go-redis/redis/v8"
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
	Weather  string
}

// главная страница с товарами
func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	userName, _ := auth.GetUserName(r)
	cacheKey := "products_list"

	cachedData, err := getFromCache(cacheKey)
	// ключ не найден в кэше, выполняем запрос к основной БД
	if err == redis.Nil {
		products, err := getProductsFromDB()
		if err != nil {
			handleError(w, err, "Error fetching products from database")
			return
		}

		if err := saveToCache(w, cacheKey, products); err != nil {
			log.Printf("Failed to save data in Redis: %v", err)
		}

		renderNewPageDataProd(w, products, userName)

		log.Println("[Redis] The data is not in Redis. Load from portgres database.")

	} else if err != nil {
		http.Error(w, "Error fetching data from Redis: %v", http.StatusInternalServerError)
		log.Printf("Failed to get data from Redis: %v", err)
	} else {
		// данные кеша есть в Редисе
		var products []Product
		if err = json.Unmarshal([]byte(cachedData), &products); err != nil {
			http.Error(w, "Error marshaling products data", http.StatusInternalServerError)
			return
		}

		renderNewPageDataProd(w, products, userName)

		log.Println("[Redis] The data from Redis has been uploaded!")
	}
}

// getting data from Redis
func getFromCache(cacheKey string) (string, error) {
	return database.Rdb.Get(database.Rdb.Context(), cacheKey).Result()
}

func getProductsFromDB() ([]Product, error) {
	db, err := database.Connect()
	if err != nil {
		log.Println("Error connecting to the databaseз", err)
		return nil, err
	}
	defer db.Close()

	// Выполнение SQL запроса для выборки всех товаров из таблицы "products"
	query := `SELECT name, price, id, image_url FROM products`
	rows, err := db.Query(query)
	if err != nil {
		log.Println("Error executing query:", err)
		return nil, err
	}
	defer rows.Close()

	// создание списка для хранения товаров
	var products []Product
	// считывание данных о товарах из результатов запроса
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.Name, &product.Price, &product.ID, &product.ImageURL); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}
		// Добавление продуктов в список
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating rows:", err)
		return nil, err
	}

	return products, nil
}

func handleError(w http.ResponseWriter, err error, message string) {
	http.Error(w, message, http.StatusInternalServerError)
	log.Println(err)
}

// сохраняем данные в Redis
func saveToCache(w http.ResponseWriter, cacheKey string, products []Product) error {
	productsJSON, err := json.Marshal(products)
	if err != nil {
		http.Error(w, "Error marshaling products data", http.StatusInternalServerError)
		return err
	}
	return database.Rdb.Set(database.Rdb.Context(), cacheKey, productsJSON, 10*time.Second).Err()
}

func renderNewPageDataProd(w http.ResponseWriter, products []Product, userName string) {
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
