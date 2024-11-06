package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/KrepkiyOrex/inter_store/internal/auth"
	"github.com/KrepkiyOrex/inter_store/internal/database"
	"github.com/KrepkiyOrex/inter_store/internal/utils"
	log "github.com/sirupsen/logrus"

	"github.com/go-redis/redis/v8"
)

type PageData struct {
	User
	ProductsData
	UserCookie
}

type Product struct {
	ID         int
	MongoID    string
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

	cachedData, err := getFromRedisCache(cacheKey)
	if err == redis.Nil {
		// ключ не найден в кэше, выполняем запрос к основной БД
		log.Info("[Redis] Cache miss. Fetching data from the database.")
		products, err := getProductsFromPostgre()
		if err != nil {
			handleError(w, err, "Error fetching products from database")
			return
		}

		if err := saveToCache(w, cacheKey, products); err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"cacheKey": cacheKey,
			}).Warn("Failed to save data in Redis")
		}

		renderNewPageDataProd(w, products, userName)


	} else if err != nil {
		// error while retrieving data from Redis
		handleError(w, err, "Error fetching data from Redis")
		log.WithFields(log.Fields{
			"error": err,
			"cacheKey": cacheKey,
		}).Error("Failed to get data from Redis")
	} else {
		// данные кеша есть в Редисе
		var products []Product
		if err = json.Unmarshal([]byte(cachedData), &products); err != nil {
			handleError(w, err, "Error unmarshaling products data from Redis")
			return
		}

		log.Info("[Redis] Cache hit. Data loaded successfully from Redis.")
		renderNewPageDataProd(w, products, userName)
	}
}

// getting data from Redis
func getFromRedisCache(cacheKey string) (string, error) {
	return database.Rdb.Get(database.Rdb.Context(), cacheKey).Result()
}

// get products from postgreSQL
func getProductsFromPostgre() ([]Product, error) {
	db, err := database.Connect()
	if err != nil {
		log.Error("Error connecting to the database", err)
		return nil, err
	}
	defer db.Close()

	// Выполнение SQL запроса для выборки всех товаров из таблицы "products"
	query := `SELECT name, price, id, image_url, mongo_id FROM products`
	rows, err := db.Query(query)
	if err != nil {
		log.Error("Error executing query:", err)
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.Name, &product.Price, &product.ID, &product.ImageURL, &product.MongoID); err != nil {
			log.Error("Error scanning row:", err)
			return nil, err
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		log.Error("Error iterating rows:", err)
		return nil, err
	}

	log.Info("Products successfully retrieved from database")
	return products, nil

}

func handleError(w http.ResponseWriter, err error, message string) {
	http.Error(w, message, http.StatusInternalServerError)
	log.WithFields(log.Fields{
		"error":   err,
		"message": message,
	}).Error("Internal server error")
}

// сохраняем данные в Redis
func saveToCache(w http.ResponseWriter, cacheKey string, products []Product) error {
	productsJSON, err := json.Marshal(products)
	if err != nil {
		http.Error(w, "Error marshaling products data", http.StatusInternalServerError)
		return err
	}
	err = database.Rdb.Set(database.Rdb.Context(), cacheKey, productsJSON, 10*time.Second).Err()
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"cachKey": cacheKey,
		}).Error("Failed to save data in Redis")
	}

	return err
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
        log.Error("Invalid request method")
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    db, err := database.Connect()
    if err != nil {
        log.Error("Error connecting to the database:", err)
        http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
        return
    }
    defer db.Close()

    var requestData struct {
        ProductID int    `json:"product_id"`
        Quantity  int    `json:"quantity"`
        MongoID   string `json:"mongo_id"`
    }

    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
        log.Error("Invalid request body:", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    userName, err := auth.GetUserName(r)
    if err != nil {
        log.Error("User not authenticated:", err)
        http.Error(w, "User not authenticated", http.StatusUnauthorized)
        return
    }

    var userID int
    err = db.QueryRow("SELECT user_id FROM users WHERE username = $1", userName).Scan(&userID)
    if err != nil {
        log.Error("User not found:", err)
        http.Error(w, "User not found", http.StatusInternalServerError)
        return
    }

	fmt.Println("	BUG		Before:", requestData)

    _, err = db.Exec(`
        INSERT INTO carts (user_id, product_id, quantity, mongo_id) 
        VALUES ($1, $2, $3, $4) 
        ON CONFLICT (user_id, product_id) 
        DO UPDATE SET quantity = carts.quantity + EXCLUDED.quantity`,
        userID, requestData.ProductID, requestData.Quantity, requestData.MongoID,
    )

	fmt.Println("	BUG		After:", requestData)
	
    if err != nil {
        log.Error("Error adding product to cart:", err)
        http.Error(w, "Error adding product to cart", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

