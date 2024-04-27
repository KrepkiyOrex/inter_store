package auth

import (
	"First_internet_store/internal/database"
	"fmt"
	"log"
	"net/http"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Подключаемся к базе данных PostgreSQL
	db, err := database.Connect()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		log.Fatal("Error connecting to the database:", err)
		return
	}
	defer db.Close()

	// Получаем данные, отправленные пользователем с формы регистрации
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
		log.Fatal("Error parsing form data:", err)
		return
	}

	// Извлекаем данные из формы
	userName := r.Form.Get("username")
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	// Сохраняем данные пользователя в базе данных
	_, err = db.Exec(
		"INSERT INTO users (username, email, password)"+
			"VALUES ($1, $2, $3)",
		userName, email, password)

	if err != nil {
		http.Error(w, "Error saving user data to the database", http.StatusInternalServerError)
		log.Fatal("Error saving user data to the database:", err)
		return
	}

	// Выводим сообщение о успешной регистрации
	fmt.Fprintf(w, "User %s successfully registered!", userName)

	log.Printf("User %s added to the database", userName)
}

// Обработчик для отображения HTML-страницы регистрации
func ShowRegistrationPage(w http.ResponseWriter, r *http.Request) {
	link := "/home/mrx/Documents/Programm Go/Results/2024.04.19_First_internet_store/First_internet_store/web/views/register.html"
	http.ServeFile(w, r, link)
}
