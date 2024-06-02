package models

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"First_internet_store/internal/auth"
	"First_internet_store/internal/database"
)

// Функция для получения имени пользователя из куки
func getUserName(r *http.Request) (string, error) {
	// Получаем значение куки с именем пользователя
	cookie, err := r.Cookie("userName")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// type PageData struct {
// 	UserName string
// }

// func renderTemplate(w http.ResponseWriter, data PageData, tmpl ...string) {
// 	template, err := template.ParseFiles(tmpl...)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	err = template.Execute(w, data)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }

type Product struct {
	Name  string
	Price int
	ID    int
}

type PageData struct {
	UserName string
	Products []Product
}

func renderTemplate(w http.ResponseWriter, data PageData, tmpl ...string) {
	template, err := template.ParseFiles(tmpl...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = template.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Account(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		UserName: "YourUserName", // Замените на имя пользователя
		// Products: products,
	}

	renderTemplate(w, data,
		"web/views/account.html",
		"web/css/navigation.html")
}

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

	data := PageData{
		UserName: "YourUserName", // Замените на имя пользователя
		Products: products,
	}

	// renderTemplate(w, data, "web/views/products.html",
	// 	"web/css/navigation.html", "web/css/style_products.html")
	renderTemplate(w, data,
		"web/views/products.html",
		"web/css/navigation.html",
		"web/css/style.html")
	// renderTemplate(w, data, "web/views/products.html")
}

// Главная страница с товарами
// func ProductsHandler(w http.ResponseWriter, r *http.Request) {
// 	// userName, err := getUserName(r)
// 	// if err != nil {
// 	// 	// Куки не найдено, показываем форму входа
// 	// 	// renderTemplate(w, PageData{}, "web/views/*.html")
// 	// 	renderTemplate(w, PageData{}, "web/views/products.html", "web/views/body.html")
// 	// 	// http.Redirect(w, r, "/login", http.StatusFound)
// 	// 	return
// 	// }

// 	// Connect to the database
// 	db, err := database.Connect()
// 	if err != nil {
// 		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
// 		log.Println("Error connecting to the database")
// 		return
// 	}

// 	// Выполнение SQL запроса для выборки всех товаров из таблицы "products"
// 	rows, err := db.Query("SELECT name, price, id FROM products")
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	defer rows.Close()

// 	// Создание списка для хранения товаров
// 	var products []struct {
// 		Name  string
// 		Price int
// 		ID    int
// 	}

// 	// Считывание данных о товарах из результатов запроса
// 	for rows.Next() {
// 		var product struct {
// 			Name  string
// 			Price int
// 			ID    int
// 		}
// 		if err := rows.Scan(&product.Name, &product.Price, &product.ID); err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		// Добавление продуктов в список
// 		products = append(products, product)
// 	}

// 	if err := rows.Err(); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	/*
// 		Забей на css! сделай уже через шаблоны и не мучайся
// 	*/

// 	// Загружаем шаблон страницы товаров и передаем ему данные о товарах
// 	// link := "/home/mrx/Documents/Programm Go/Results/2024.04.19_First_internet_store/First_internet_store/web/views/products.html"
// 	// link := "./web/views/products.html"
// 	link1 := "web/views/products.html"
// 	link2 := "web/css/navigation.html"
// 	link3 := "web/css/style.html"
// 	// link3 := "web/views/style.html"

// 	userName, err := getUserName(r)
// 	data := PageData{UserName: userName}
// 	// renderTemplate(w, PageData{}, link1, link2)
// 	// renderTemplate(w, data, link1, link2)

// 	if err != nil {
// 		// Куки не найдено, показываем форму входа
// 		// renderTemplate(w, PageData{}, "template/index.html", "template/body.html") // ParseFiles
// 		renderTemplate(w, PageData{}, link1, link2, link3) // ParseFiles

// 		// http.Redirect(w, r, "/login", http.StatusFound)
// 		return
// 	}
// 	renderTemplate(w, data, link1, link2, products) // products

// 	// tmpl := template.Must(template.ParseFiles(link1, link2)) // через Must рендж пашет. WTF
// 	// // tmpl, err := template.ParseFiles(link1, link2) // через Must рендж пашет. WTF
// 	// // template.ParseFiles(link1, link2) // через Must рендж пашет. WTF
// 	// err = tmpl.Execute(w, products) // вот из-за чего рендж пашет!
// 	// if err != nil {
// 	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
// 	// 	return
// 	// }

// 	/*
// 		template.Must возможно нехватает
// 	*/

// 	// data := PageData{UserName: userName}
// 	// renderTemplate(w, data, "web/views/products.html", "web/views/body.html")
// 	// renderTemplate(w, data, "web/views/products.html")
// 	// renderTemplate(w, data, "web/views/body_tmpl.html")
// 	// renderTemplate(w, data, "web/views/products.html")
// 	// renderTemplate(w, data, "/web/views/*.html")
// 	// renderTemplate(w, data, "/views/*.html")
// 	// renderTemplate(w, data, "views/*.html")

// 	log.Print("Handling request for products")
// }

// Обработчик для добавления товара в корзину
func AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем токен из заголовка запроса
	tokenString := auth.ExtractToken(r)
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем информацию о пользователе из токена
	user, err := auth.GetUserFromToken(tokenString)
	if err != nil {
		http.Error(w, "Failed to get user from token", http.StatusInternalServerError)
		return
	}

	// Парсим данные товара из запроса
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
		return
	}
	// productName := r.Form.Get("product_name")
	// productPrice := r.Form.Get("product_price")

	productIDStr := r.Form.Get("product_id")
	productID, err := strconv.Atoi(productIDStr) // Преобразуем строку в целое число
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Добавляем товар в корзину пользователя в базе данных
	// err = addToCart(user.ID, productName, productPrice)
	err = addToCart(user.ID, productID)
	if err != nil {
		http.Error(w, "Error adding product to cart", http.StatusInternalServerError)
		return
	}

	// Если все прошло успешно, отправляем клиенту подтверждение
	fmt.Fprintf(w, "Product %s added to cart successfully!", productIDStr)
}

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

// Обработчик для всех запросов
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Если запрос POST, обрабатываем аутентификацию
		userName := r.FormValue("userName")

		// Создаем новую куку с именем пользователя
		cookie := http.Cookie{
			Name:  "userName",
			Value: userName,
		}

		// Устанавливаем куку в ответ
		http.SetCookie(w, &cookie)

		// Перенаправляем пользователя на главную страницу
		http.Redirect(w, r, "/hello", http.StatusFound)
	} else {
		// Если запрос GET, отображаем главную страницу
		// Получаем значение куки с именем пользователя
		userNameCookie, err := r.Cookie("userName")
		if err == nil {
			// Если куки существует, отображаем имя пользователя на странице
			userName := userNameCookie.Value
			// Используем шаблон для вставки имени пользователя на страницу
			tmpl := template.Must(template.New("index").Parse(`
                <html>
                <body>
                    <p>Привет, test {{.UserName}}!</p>
                    <!-- Форма для аутентификации -->
                    <form action="/login" method="post">
                        <input type="text" name="userName">
                        <input type="submit" value="Войти">
                    </form>
                </body>
                </html>
            `))
			tmpl.Execute(w, map[string]interface{}{
				"UserName": userName,
			})
		} else {
			// Если куки не существует, отображаем страницу без имени пользователя и форму для аутентификации
			http.ServeFile(w, r, "login.html")
		}
	}
}

func ViewCartHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "View cart page")
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	// link := "web/views/list.html"
	// http.ServeFile(w, r, link)

	// ==================================================

	userName, err := getUserName(r)
	if err != nil {
		// Куки не найдено, показываем форму входа
		renderTemplate(w, PageData{}, "web/views/list.html")
		// http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	data := PageData{UserName: userName}
	renderTemplate(w, data, "web/views/list.html")
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
