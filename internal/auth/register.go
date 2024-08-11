package auth

import (
	"First_internet_store/internal/database"
	"First_internet_store/internal/utils"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// для получения имени пользователя из куки
func GetUserName(r *http.Request) (string, error) {
	cookie, err := r.Cookie("userName")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// проверка на дупликаты аккаунтов
func checkDuplicateAccounts(db *database.DB, email string) (bool, error) {
	var existingEmail string
	err := db.QueryRow("SELECT email FROM users WHERE email = $1", email).Scan(&existingEmail)
	if err == sql.ErrNoRows {
		return false, nil // Нет дубликата
	} else if err != nil {
		return false, err // Другие ошибки
	}
	return true, nil // Дубликат найден
}

// регистрирует пользователя
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		utils.RenderTemplate(w, PageData{},
			"web/html/register.html",
			"web/html/navigation.html")
		return
	}

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

	// извлекаем данные из формы
	userName := r.Form.Get("username")
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	if !fieldValidate(userName, email, password, w) {
		return
	}

	// Проверка дубликатов
	duplicate, err := checkDuplicateAccounts(db, email)
	if err != nil {
		http.Error(w, "Error checking for duplicate accounts", http.StatusInternalServerError)
		log.Fatal("Error checking for duplicate accounts:", err)
		return
	}

	if duplicate {
		data := PageData{}.newPageData(userName, email, "Email already registered")

		utils.RenderTemplate(w, data,
			"web/html/register.html",
			"web/html/navigation.html")
		return
	}

	// Сохраняем данные пользователя в базе данных "users"
	var userID int
	err = db.QueryRow(`
		INSERT INTO users (username, email, password) 
			VALUES ($1, $2, $3) 
			RETURNING user_id`,
		userName, email, password).Scan(&userID)
	if err != nil {
		http.Error(w, "Error saving user data to the database", http.StatusInternalServerError)
		log.Fatal("Error saving user data to the database:", err)
		return
	}

	// устанавливаем ник в куки
	SetCookie(w, "userName", userName, time.Now().Add(24*time.Hour))

	// устанавливаем ID пользователя в куки
	SetCookie(w, "userID", strconv.Itoa(userID), time.Now().Add(24*time.Hour))

	log.Printf("User %s added to the database with ID %d", userName, userID)

	// Перенаправляем пользователя на страницу аккаунта
	http.Redirect(w, r, "/account", http.StatusSeeOther)
}

type PageData struct {
	User          User
	PersonDetails PersonDetails
	UserError     UserError
}

// для вывода ошибок для пользователя
type UserError struct {
	ErrorMessage string
}

// отображение информации о пользователе, а также любую ошибку
func (pd PageData) newPageData(userName, email, errMsg string) PageData {
	return PageData{
		User: User{
			UserName: userName,
			Email:    email,
		},
		UserError: UserError{ErrorMessage: errMsg},
	}
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

func validateAndRender(w http.ResponseWriter, field, fieldName, userName, email string) bool {
	if strings.TrimSpace(field) == "" {
		utils.RenderTemplate(w, PageData{
			User: User{
				UserName: userName,
				Email:    email},
			UserError: UserError{ErrorMessage: fieldName + " cannot be empty"},
		}, "web/html/register.html", "web/html/navigation.html")
		return false
	}
	return true
}

// Обработчик для отображения HTML-страницы регистрации
func ShowRegistrationPage(w http.ResponseWriter, r *http.Request) {
	// link := "web/html/register.html"
	// http.ServeFile(w, r, link)

	utils.RenderTemplate(w, PageData{},
		// utils.RenderTemplate(w, PageData{UserError: UserError{}},
		"web/html/register.html",
		"web/html/navigation.html")
}
