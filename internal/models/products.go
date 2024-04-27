package models

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"First_internet_store/internal/database"

	"First_internet_store/internal/utils"
)

// Главная страница с товарами
func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	// Connect to the database
	db, err := database.Connect()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		log.Println("Error connecting to the database")
		return
	}

	// Выполнение SQL запроса для выборки всех товаров из таблицы "products"
	rows, err := db.Query("SELECT name, price, id FROM products")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Создание списка для хранения товаров
	var products []struct {
		Name  string
		Price int
		ID    int
	}

	// Считывание данных о товарах из результатов запроса
	for rows.Next() {
		var product struct {
			Name  string
			Price int
			ID    int
		}
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

	// Загружаем шаблон страницы товаров и передаем ему данные о товарах
	link := "/home/mrx/Documents/Programm Go/Results/2024.04.19_First_internet_store/First_internet_store/web/views/products.html"
	tmpl := template.Must(template.ParseFiles(link))
	err = tmpl.Execute(w, products)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Print("Handling request for products")
}

// Обработчик для добавления товара в корзину
func AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем идентификатор товара из запроса (здесь предполагается, что у вас есть идентификатор товара)
	productID, err := strconv.Atoi(r.FormValue("product_id"))
	if err != nil {
		http.Error(w, "Invalid error ID", http.StatusBadRequest)
		return
	}

	// Получаем сессию пользователя
	session, err := utils.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем или создаем массив для товаров в корзине
	cart, ok := session.Values["cart"].([]string)
	if !ok {
		cart = []string{}
	}

	// Добавляем идентификатор товара в корзину
	cart = append(cart, strconv.Itoa(productID))

	// Сохраняем обновленную корзину в сессии
	session.Values["cart"] = cart
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Перенаправляем пользователя обратно на страницу с товарами
	http.Redirect(w, r, "/products", http.StatusSeeOther)
}

func ViewCartHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "View cart page")
}
