package models

import (
	"fmt"
	"log"
	"net/http"

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
	// var products []utils.Product
	var products []Product

	// Считывание данных о товарах из результатов запроса
	for rows.Next() {
		// var product utils.Product
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

	// data := UserCookie{
	// 	UserName: "userName", // Замените на имя пользователя
	// }

	// Выгрузка данных продуктов
	// data := ProductsData{
	// 	Products: products,
	// }

	// для имени из куки
	// var userName string

	// cookie, err := r.Cookie("userName")
	// if err == nil {
	// 	userName = cookie.Value
	// }

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
	ID    int
	Name  string
	Price float64
}

type ProductsData struct {
	Products []Product
}

type UserCookie struct {
	UserName string
}

/*
	Переделай здесь ExtractToken и GetUserFromToken для добавления в корзину, 
	после того, как сделаешь каждому пользователю ЛК. ЧИТАЙ В ФАЙЛЕ 
	АРХИТЕКТУРА на 92 строчке про это. 
 */
// Обработчик для добавления товара в корзину
// func AddToCartHandler(w http.ResponseWriter, r *http.Request) {
// 	// Извлекаем токен из заголовка запроса
// 	tokenString := auth.ExtractToken(r)
// 	if tokenString == "" {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	// Получаем информацию о пользователе из токена
// 	user, err := auth.GetUserFromToken(tokenString)
// 	if err != nil {
// 		http.Error(w, "Failed to get user from token", http.StatusInternalServerError)
// 		return
// 	}

// 	// Парсим данные товара из запроса
// 	err = r.ParseForm()
// 	if err != nil {
// 		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
// 		return
// 	}
// 	// productName := r.Form.Get("product_name")
// 	// productPrice := r.Form.Get("product_price")

// 	productIDStr := r.Form.Get("product_id")
// 	productID, err := strconv.Atoi(productIDStr) // Преобразуем строку в целое число
// 	if err != nil {
// 		http.Error(w, "Invalid product ID", http.StatusBadRequest)
// 		return
// 	}

// 	// Добавляем товар в корзину пользователя в базе данных
// 	// err = addToCart(user.ID, productName, productPrice)
// 	err = addToCart(user.ID, productID)
// 	if err != nil {
// 		http.Error(w, "Error adding product to cart", http.StatusInternalServerError)
// 		return
// 	}

// 	// Если все прошло успешно, отправляем клиенту подтверждение
// 	fmt.Fprintf(w, "Product %s added to cart successfully!", productIDStr)
// }

func addToCart(userID int, productID int) error {
	// Подключаемся к базе данных
	db, err := database.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	// Выполняем запрос для добавления товара в корзину
	_, err = db.Exec("INSERT INTO order_items (user_id, product_id) VALUES ($1, $2)", userID, productID)
	if err != nil {
		return err
	}

	return nil
}

// DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED?
// DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED?
// Обработчик для всех запросов
// func HelloHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == "POST" {
// 		// Если запрос POST, обрабатываем аутентификацию
// 		userName := r.FormValue("userName")

// 		// Создаем новую куку с именем пользователя
// 		cookie := http.Cookie{
// 			Name:  "userName",
// 			Value: userName,
// 		}

// 		// Устанавливаем куку в ответ
// 		http.SetCookie(w, &cookie)

//			// Перенаправляем пользователя на главную страницу
//			http.Redirect(w, r, "/hello", http.StatusFound)
//		} else {
//			// Если запрос GET, отображаем главную страницу
//			// Получаем значение куки с именем пользователя
//			userNameCookie, err := r.Cookie("userName")
//			if err == nil {
//				// Если куки существует, отображаем имя пользователя на странице
//				userName := userNameCookie.Value
//				// Используем шаблон для вставки имени пользователя на страницу
//				tmpl := template.Must(template.New("index").Parse(`
//	                <html>
//	                <body>
//	                    <p>Привет, test {{.UserName}}!</p>
//	                    <!-- Форма для аутентификации -->
//	                    <form action="/login" method="post">
//	                        <input type="text" name="userName">
//	                        <input type="submit" value="Войти">
//	                    </form>
//	                </body>
//	                </html>
//	            `))
//				tmpl.Execute(w, map[string]interface{}{
//					"UserName": userName,
//				})
//			} else {
//				// Если куки не существует, отображаем страницу без имени пользователя и форму для аутентификации
//				http.ServeFile(w, r, "login.html")
//			}
//		}
//	}
//

func EditProduct(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что метод запроса POST
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	// Получаем данные из формы
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Получаем ID товара и новые данные
	productID := r.FormValue("product_id")
	productName := r.FormValue("product_name")
	productPrice := r.FormValue("product_price")

	// Здесь можно обновить данные товара в базе данных
	// Например, выполнить SQL-запрос к базе данных и обновить соответствующий товар

	// Выводим сообщение об успешном обновлении (это можно изменить по вашему усмотрению)
	fmt.Fprintf(w, "Товар с ID %s успешно обновлен. Новое название: %s, новая цена: %s", productID, productName, productPrice)
}

func ViewCartHandler(w http.ResponseWriter, r *http.Request) {
	// для имени из куки
	// var userName string

	// cookie, err := r.Cookie("userName")
	// if err == nil {
	// 	userName = cookie.Value
	// }

	userName, _ := auth.GetUserName(r)

	data := PageData{
		UserCookie: UserCookie{
			UserName: userName,
		},
	}

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
		// http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	data := UserCookie{UserName: userName}
	utils.RenderTemplate(w, data,
		"web/html/list.html",
		"web/html/navigation.html")
}

// Обработчик для добавления товара в корзину
// func AddToCartHandler(w http.ResponseWriter, r *http.Request) {
// 	// Получаем идентификатор товара из запроса (здесь предполагается, что у вас есть идентификатор товара)
// 	productID, err := strconv.Atoi(r.FormValue("product_id"))
// 	if err != nil {
// 		http.Error(w, "Invalid error ID", http.StatusBadRequest)
// 		return
// 	}

// 	// Получаем сессию пользователя
// 	session, err := utils.Store.Get(r, "session-name")
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Получаем или создаем массив для товаров в корзине
// 	cart, ok := session.Values["cart"].([]string)
// 	if !ok {
// 		cart = []string{}
// 	}

// 	// Добавляем идентификатор товара в корзину
// 	cart = append(cart, strconv.Itoa(productID))

// 	// Сохраняем обновленную корзину в сессии
// 	session.Values["cart"] = cart
// 	err = session.Save(r, w)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Перенаправляем пользователя обратно на страницу с товарами
// 	http.Redirect(w, r, "/products", http.StatusSeeOther)
// }
