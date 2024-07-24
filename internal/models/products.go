package models

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"First_internet_store/internal/auth"
	"First_internet_store/internal/database"
	"First_internet_store/internal/utils"
)

// Функция для получения имени пользователя из куки
// func getUserName(r *http.Request) (string, error) {
// 	// Получаем значение куки с именем пользователя
// 	cookie, err := r.Cookie("userName")
// 	if err != nil {
// 		return "", err
// 	}
// 	return cookie.Value, nil
// }

// user profile
// func Account(w http.ResponseWriter, r *http.Request) {
// 	// Извлекаем куку
// 	cookie, err := r.Cookie("userName")
// 	userName := "" // Значение по умолчанию, если кука не установлена
// 	if err == nil {
// 		userName = cookie.Value
// 	}

// 	data := UserCookie{
// 		UserName: userName,
// 	}

// 	utils.RenderTemplate(w, data,
// 		// renderTemplate(w, data,
// 		"web/html/account.html",
// 		"web/html/navigation.html")
// }

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
	rows, err := db.Query("SELECT name, price, id FROM products")
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

		if err := rows.Scan(&product.Name, &product.Price, &product.ID); err != nil {
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

	// 12
	userName, _ := auth.GetUserName(r) // 32

	data := PageData{
		ProductsData: ProductsData{
			Products: products,
		},
		UserCookie: UserCookie{
			UserName: userName,
		},
	}

	utils.RenderTemplate(w, data,
		"web/html/products.html",
		"web/html/navigation.html",
	)
}

type PageData struct {
	ProductsData
	UserCookie
}

type Product struct {
	ID         int
	Name       string
	Price      float64
	Quantity   int
	TotalPrice float64 // Добавляем поле для общей стоимости
}

type ProductsData struct {
	Products []Product
	TotalSum float64 // Добавляем поле для общей суммы
}

type UserCookie struct {
    UserID   int
    UserName string
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

// корзина добавленых заказов, перед оплатой
func ViewCartHandler(w http.ResponseWriter, r *http.Request) {
	// Подключение к базе данных
	db, err := database.Connect()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		log.Println("Error connecting to the database")
		return
	}
	defer db.Close()

	// Получение идентификатора пользователя из куки
	cookie, err := r.Cookie("userID")
	if err != nil || cookie.Value == "" {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	// Выполнение SQL запроса для получения данных из корзины и продуктов
	query := `
		SELECT p.id, p.name, p.price, c.quantity 
		FROM carts c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id = $1` // Используем $1 для параметра

	rows, err := db.Query(query, userID) // Пример с user_id = 59
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Чтение результатов запроса
	var products []Product
	var totalSum float64
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Quantity); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		product.TotalPrice = product.Price * float64(product.Quantity) // Приведение Quantity к float64
		products = append(products, product)
		totalSum += product.TotalPrice // Суммируем общую стоимость каждого продукта
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получение имени пользователя (пример)
	userName, _ := auth.GetUserName(r) // Пример получения имени пользователя

	// Подготовка данных для шаблона
	data := PageData{
		ProductsData: ProductsData{
			Products: products,
			TotalSum: totalSum, // Передаем общую сумму в шаблон
		},
		UserCookie: UserCookie{
			UserName: userName,
		},
	}

	// Рендеринг шаблона
	utils.RenderTemplate(w, data,
		"web/html/cart.html",
		"web/html/navigation.html",
	)
}

// Содержимое вашей корзины
func ListHandler(w http.ResponseWriter, r *http.Request) {
	userName, err := auth.GetUserName(r)
	if err != nil {
		// Куки не найдено, показываем форму входа
		utils.RenderTemplate(w, UserCookie{},
			"web/html/list.html",
			"web/html/navigation.html")
		return
	}

	data := UserCookie{UserName: userName}
	utils.RenderTemplate(w, data,
		"web/html/list.html",
		"web/html/navigation.html")
}
