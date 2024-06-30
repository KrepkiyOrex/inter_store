package auth

import (
	"First_internet_store/internal/database"
	"First_internet_store/internal/utils"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// Функция для получения имени пользователя из куки
func GetUserName(r *http.Request) (string, error) {
	// Получаем значение куки с именем пользователя
	cookie, err := r.Cookie("userName")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// if r.Method == http.MethodGet {
	// 	utils.RenderTemplate(w, PageData{})
	// 	return
	// }

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

	// test duplicate
	// test duplicate
	// test duplicate

	// validate
	if !fieldValidate(userName, email, password, w) {
		return
	}

	// Сохраняем данные пользователя в базе данных "users"
	_, err = db.Exec(
		"INSERT INTO users (username, email, password)" +
			"VALUES ($1, $2, $3)",
		userName, email, password)

	if err != nil {
		http.Error(w, "Error saving user data to the database", http.StatusInternalServerError)
		log.Fatal("Error saving user data to the database:", err)
		return
	}

	// Выводим сообщение о успешной регистрации
	fmt.Fprintf(w, "User %s successfully registered!", userName)

	// добавь сюда страницу с радостями, что клиент зарегистрировался
	// добавь сюда страницу с радостями, что клиент зарегистрировался
	// добавь сюда страницу с радостями, что клиент зарегистрировался

	log.Printf("User %s added to the database", userName)
}

type PageData struct {
	UserError
	Username string
	Email    string
}

// проверка на пустые поля при регистрации
func fieldValidate(userName, email, password string, w http.ResponseWriter) bool {
	if !validateAndRender(w, userName, "Username", userName, email) {
		return false
	}
	if !validateAndRender(w, email, "Email", userName, email) {
		return false
	}
	if !validateAndRender(w, password, "Password", userName, email) {
		return false
	}
	return true
}

// проверка на пустые поля при регистрации
func validateAndRender(w http.ResponseWriter, field, fieldName, userName, email string) bool {
	if strings.TrimSpace(field) == "" {
		utils.RenderTemplate(w, PageData{
			UserError: UserError{ErrorMessage: fieldName + " cannot be empty"},
			Username:  userName,
			Email:     email,
		}, "web/html/register.html", "web/html/navigation.html")
		return false
	}
	return true
}

// func dataNotEmpty(userName, email, password string, w http.ResponseWriter) {
// 	// if userName == "" || email == "" || password == "" {
// 	// 	return fmt.Errorf("data cannot be empty")
// 	// 	// log.Fatal("data cannot be empty")
// 	// 	// return false
// 	// }
// 	// // return true
// 	// return nil

// 	if strings.TrimSpace(userName) == "" {
// 		utils.RenderTemplate(w, PageData{
// 			UserError: UserError{ErrorMessage: "Username cannot be empty"},
// 			Username:  userName, Email: email},
// 			"web/html/register.html", "web/html/navigation.html")
// 		return
// 	}
// 	if strings.TrimSpace(email) == "" {
// 		utils.RenderTemplate(w, PageData{
// 			UserError: UserError{ErrorMessage: "Email cannot be empty"},
// 			Username:  userName,
// 			Email:     email},
// 			"web/html/register.html", "web/html/navigation.html")
// 		return
// 	}
// 	if strings.TrimSpace(password) == "" {
// 		utils.RenderTemplate(w, PageData{
// 			UserError: UserError{ErrorMessage: "Password cannot be empty"},
// 			Username:  userName,
// 			Email:     email},
// 			"web/html/register.html", "web/html/navigation.html")
// 		return
// 	}
// }

// Обработчик для отображения HTML-страницы регистрации
func ShowRegistrationPage(w http.ResponseWriter, r *http.Request) {
	// link := "web/html/register.html"
	// http.ServeFile(w, r, link)

	utils.RenderTemplate(w, PageData{},
		// utils.RenderTemplate(w, PageData{UserError: UserError{}},
		"web/html/register.html",
		"web/html/navigation.html")
}
