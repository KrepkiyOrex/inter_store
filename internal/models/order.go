package models

import (
	"First_internet_store/internal/auth"
	"First_internet_store/internal/database"
	"First_internet_store/internal/utils"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Ваша функция для рендеринга шаблона
func UserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	// Получение ID пользователя из куки
	cookieID, err := r.Cookie("userID")
	if err != nil {
		http.Error(w, "Invalid userID", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(cookieID.Value)
	if err != nil {
		http.Error(w, "Invalid userID format", http.StatusBadRequest)
		return
	}

	log.Println("Extracted userID from cookie:", userID)

	// Получение заказов пользователя с БД
	orders, err := getOrdersForUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if orders == nil {
		orders = []Order{} // Инициализация пустого списка заказов, если nil
	}

	userName, _ := auth.GetUserName(r)

	data := OrderPageData{}.newOrderPageData(orders, userID, userName)

	utils.RenderTemplate(w, data,
		"web/html/orders.html",
		"web/html/navigation.html")
}

// создание данных для страницы продуктов и куки пользователя
func (opd OrderPageData) newOrderPageData(orders []Order, userID int, userName string) OrderPageData {
	return OrderPageData{
		OrdersDate: OrdersDate{
			Orders: orders,
		},
		UserCookie: UserCookie{
			UserID:   userID,
			UserName: userName,
		},
	}
}

type OrderPageData struct {
	OrdersDate
	UserCookie
}

// Структура для предоставления заказа
type Order struct {
	UserID             int
	TotalAmount        float32
	OrderDate          time.Time
	FormattedOrderDate string // поле для форматированной даты
	PaymentStatus      string
	ShippingAddress    string
}

type OrdersDate struct {
	Orders []Order
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

// обновление счетчика в корзине во время оформления

// обновление счетчика в корзине во время оформления
func UpdateCartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := utils.GetUserIDFromCookie(r)
	if err != nil {
		http.Error(w, "User ID not found: "+err.Error(), http.StatusUnauthorized)
		return
	}

	productIDStr := r.FormValue("product_id")
	quantityStr := r.FormValue("quantity")
	if productIDStr == "" || quantityStr == "" {
		http.Error(w, "Product ID or quantity missing", http.StatusBadRequest)
		return
	}

	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		http.Error(w, "Invalid quantity: "+err.Error(), http.StatusBadRequest)
		return
	}

	db, err := database.Connect()
	if err != nil {
		http.Error(w, "Error connecting to the database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec(`
        INSERT INTO carts (user_id, product_id, quantity)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id, product_id)
        DO UPDATE SET quantity = EXCLUDED.quantity
    `, userID, productID, quantity)

	if err != nil {
		http.Error(w, "Database update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// depr
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

// Для получения купленых заказов пользователя из БД orders
func getOrdersForUser( /* db *sql.DB, */ userId int) ([]Order, error) {
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
		return nil, err
	}
	defer db.Close()

	query := `
		SELECT user_id, total_amount, order_date, 
		payment_status, shipping_address
		FROM orders
		WHERE user_id = $1
	`
	// Выполнение запроса к базе данных
	rows, err := db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Создание списка для хранения заказов
	var orders []Order

	// Обработка результатов запроса
	for rows.Next() {
		var order Order

		// Сканирование данных заказа из строк запроса
		if err := rows.Scan(
			&order.UserID,
			&order.TotalAmount,
			&order.OrderDate,
			&order.PaymentStatus,
			&order.ShippingAddress); err != nil {
			return nil, err
		}

		// Форматирование даты и времени
		order.FormattedOrderDate = order.OrderDate.Format("2006-01-02 15:04:05")

		// Добавление заказа в список
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// Перенос с добавленых в корзину заказов, в историю оплаченых заказов
func SubmitOrderHandler(w http.ResponseWriter, r *http.Request) {
	// Подключение к базе данных
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Error connecting to the database", err)
	}
	defer db.Close()

	// Получение идентификатора пользователя из куки
	cookieID, err := r.Cookie("userID")
	if err != nil || cookieID.Value == "" {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(cookieID.Value)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	// Получение данных из формы
	address := r.FormValue("address")
	// delivery := r.FormValue("delivery")
	payment := r.FormValue("payment")

	// Извлечение данных из таблицы cart
	type CartItem struct {
		CartID    int
		UserID    int
		ProductID int
		Quantity  int
		DateAdded time.Time
	}

	var cartItems []CartItem

	query := `SELECT cart_id, user_id, product_id, quantity, date_added 
			FROM carts 
			WHERE user_id = $1`

	rows, err := db.Query(query, userID)
	if err != nil {
		http.Error(w, "Error fetching cart items", http.StatusInternalServerError)
		log.Println("Error fetching cart items:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item CartItem
		if err := rows.Scan(&item.CartID, &item.UserID, &item.ProductID, &item.Quantity, &item.DateAdded); err != nil {
			http.Error(w, "Error scanning cart items", http.StatusInternalServerError)
			log.Println("Error scanning cart items:", err)
			return
		}
		cartItems = append(cartItems, item)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating over cart items", http.StatusInternalServerError)
		log.Println("Error iterating over cart items:", err)
		return
	}

	// Подготовка данных для таблицы order
	var totalAmount float64

	for _, item := range cartItems {
		var price float64
		err = db.QueryRow("SELECT price FROM products WHERE id = $1", item.ProductID).Scan(&price)
		if err != nil {
			http.Error(w, "Error fetching product price", http.StatusInternalServerError)
			log.Println("Error fetching product price:", err)
			return
		}
		totalAmount += price * float64(item.Quantity)
	}

	orderQuery := `INSERT INTO orders 
				(user_id, order_date, total_amount, shipping_address, payment_status) 
            	VALUES ($1, $2, $3, $4, $5) RETURNING order_id`

	var orderID int
	err = db.QueryRow(orderQuery, userID, time.Now(), totalAmount, address, payment).Scan(&orderID)
	if err != nil {
		http.Error(w, "Error creating order", http.StatusInternalServerError)
		log.Println("Error creating order:", err)
		return
	}

	// Очистка таблицы carts пользователя, после оплаыт
	deleteQuery := `DELETE FROM carts WHERE user_id = $1`

	_, err = db.Exec(deleteQuery, userID)
	if err != nil {
		http.Error(w, "Error clearing cart", http.StatusInternalServerError)
		log.Println("Error clearing cart:", err)
		return
	}

	// Перенаправление пользователя на страницу /users-orders
	http.Redirect(w, r, "/users-orders", http.StatusSeeOther)
}
