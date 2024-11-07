package models

import (
	"context"
	"fmt"

	"github.com/KrepkiyOrex/inter_store/internal/auth"
	"github.com/KrepkiyOrex/inter_store/internal/database"
	"github.com/KrepkiyOrex/inter_store/internal/others"
	"github.com/KrepkiyOrex/inter_store/internal/utils"
	"github.com/KrepkiyOrex/inter_store/inventory"
	log "github.com/sirupsen/logrus"

	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// показывает заказы пользователя по ID
func UserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetCookieUserID(w, r)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		log.Error("Failed to retrieve user ID from cookie")
	}
	log.Info("Parsed User ID: ", userID)

	// Получение заказов пользователя с БД
	orders, err := getOrdersForUser(userID)
	if err != nil {
		http.Error(w, "Failed to retrieve orders", http.StatusInternalServerError)
		log.Error("Error retrieving user orders: ", err)
		return
	}

	if orders == nil {
		log.Info("No orders found for user: ", userID)
		orders = []Order{} // инициализация пустого списка заказов, если nil
	}

	userName, _ := auth.GetUserName(r)

	data := OrderPageData{}.newOrderPageData(orders, userID, userName)
	log.Info("Renderind user ordors page for user: ", userID)

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
	db, err := database.Connect()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		log.Println("Error connecting to the database")
		return
	}
	defer db.Close()

	userID, err := auth.GetCookieUserID(w, r)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
	}
	log.Println("Parsed User ID:", userID)

	// Выполнение SQL запроса для получения данных из корзины и продуктов
	query := `
		SELECT p.id, p.name, p.price, c.quantity 
		FROM carts c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id = $1`

	rows, err := db.Query(query, userID)
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
		product.TotalPrice = product.Price * float64(product.Quantity)
		products = append(products, product)
		totalSum += product.TotalPrice // Суммируем общую стоимость каждого продукта
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получение имени пользователя
	userName, _ := auth.GetUserName(r) // Пример получения имени пользователя

	renderViewCartPage(w, r, userName, products, totalSum)
}

func (pd PageData) newCartPageData(w http.ResponseWriter, userName string, products []Product, totalSum float64) PageData {
	// подготовка данных для шаблона
	return PageData{
		ProductsData: ProductsData{
			Products: products,
			TotalSum: totalSum, // передаем общую сумму товаров в шаблон
		},
		UserCookie: UserCookie{
			UserName: userName,
		},
	}
}

func renderViewCartPage(w http.ResponseWriter, r *http.Request, userName string, products []Product, totalSum float64) {
	data := PageData{}.newCartPageData(w, userName, products, totalSum)

	// Рендеринг шаблона
	utils.RenderTemplate(w, data,
		"web/html/cart.html",
		"web/html/navigation.html",
	)
}

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
// func ListHandler(w http.ResponseWriter, r *http.Request) {
// 	userName, err := auth.GetUserName(r)
// 	if err != nil {
// 		// Куки не найдено, показываем форму входа
// 		utils.RenderTemplate(w, UserCookie{},
// 			"web/html/list.html",
// 			"web/html/navigation.html")
// 		return
// 	}

// 	data := UserCookie{UserName: userName}
// 	utils.RenderTemplate(w, data,
// 		"web/html/list.html",
// 		"web/html/navigation.html")
// }

// для получения купленых заказов пользователя из БД orders
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

// перенос с добавленых в корзину заказов, в историю оплаченых заказов
func SubmitOrderHandler(w http.ResponseWriter, r *http.Request) {
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Error connecting to the database", err)
	}
	defer db.Close()

	userID, err := auth.GetCookieUserID(w, r)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
	}
	log.Println("Parsed User ID:", userID)

	// Получение данных из формы
	address := r.FormValue("address")
	// delivery := r.FormValue("delivery")
	payment := r.FormValue("payment")

	// Извлечение данных из таблицы carts
	type CartItem struct {
		CartID    int
		UserID    int
		ProductID int
		MongoID   string
		Quantity  int
		DateAdded time.Time
	}

	var cartItems []CartItem

	query := `SELECT cart_id, user_id, product_id, quantity, date_added, mongo_id
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
		if err := rows.Scan(
			&item.CartID,
			&item.UserID,
			&item.ProductID,
			&item.Quantity,
			&item.DateAdded,
			&item.MongoID); err != nil {
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

	log.Println("Mongo ID:", cartItems)

	// Массив для хранения ID продуктов и их количеств для декрементации
	for _, item := range cartItems {
		var price float64
		err = db.QueryRow("SELECT price FROM products WHERE id = $1", item.ProductID).Scan(&price)
		if err != nil {
			http.Error(w, "Error fetching product price", http.StatusInternalServerError)
			log.Println("Error fetching product price:", err)
			return
		}
		totalAmount += price * float64(item.Quantity)

		// Декрементирование товара в MongoDB через gRPC
		success, err := DecremInventoryGRPC(item.MongoID, int32(item.Quantity))
		if err != nil || !success {
			http.Error(w, "Failed to update inventory", http.StatusInternalServerError)
			log.Error("Failed to decrement inventory:", err)
			return
		}
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

	// Очистка таблицы carts пользователя, после оплаты
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

func DecremInventoryGRPC(productID string, quantity int32) (bool, error) {
	invClient, conn, err := others.NewInventoryClient()
	if err != nil {
		return false, fmt.Errorf("failed to create gRPC inventory client: %v", err)
	}
	defer conn.Close()

	// Создаём запрос с `productID` и `quantity`
	req := &inventory.RemoveInventoryRequest{
		ProductId: productID,
		Quantity:  quantity, // Количество для уменьшения
	}

	// Вызываем `RemoveInventory` на клиенте
	resp, err := invClient.RemoveInventory(context.Background(), req)
	if err != nil {
		return false, fmt.Errorf("failed to update inventory via gRPC: %v", err)
	}

	// проверяем, успешно ли выполнен запрос
	return resp.Success, nil
}
