package models

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// Структура для представления заказа
type Order struct {
	ID        int
	OrderDate time.Time
}

// Для получения заказов пользователя из базы данных
func getOrdersForUser(db *sql.DB, userId int) ([]Order, error) {
	// подготовка SQL запроса
	query := `
		SELECT id, order_date
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
		if err := rows.Scan(&order.OrderDate, &order.OrderDate); err != nil {
			return nil, err
		}
		// Добавление заказа в список
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// Заказы пользователя ??????????????????
func UserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	// Получение ID пользователя из сессии или запроса в зависимости от вашей логики
	userID := 1 // Ваша логика получения идентификатора пользователя

	// Настройка подключения к БД
	db, err := sql.Open("postgres", "user=postgres password=qwerty dbname=online_store sslmode=disable")
	if err != nil {
		fmt.Println("Error connecting to the database", err)
		return
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Println("Error pinging the database", err)
		return
	}

	// Получение заказов пользователя с БД
	orders, err := getOrdersForUser(db, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Загрузка HTML шаблона
	link := "/home/mrx/Documents/Programm Go/Results/2024.04.19_First_internet_store/First_internet_store/web/views/orders.html"
	tmpl := template.Must(template.ParseFiles(link))

	// Отправка страницы HTML с данными о заказах
	if err := tmpl.Execute(w, orders); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
