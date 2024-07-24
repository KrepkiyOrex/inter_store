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

// Функция для получения имени пользователя из куки
func GetUserName(r *http.Request) (string, error) {
	// Получаем значение куки с именем пользователя
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
// регистрирует пользователя
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		utils.RenderTemplate(w, PageData{},
			"web/html/register.html",
			"web/html/navigation.html")
		return
	}

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

	// Field validation
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
	err = db.QueryRow(
		"INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING user_id",
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

// // Обработчик для страницы приветствия
// func WelcomeHandler(w http.ResponseWriter, r *http.Request) {
// 	utils.RenderTemplate(w, PageData{},
// 		"web/html/welcome.html",
// 		"web/html/navigation.html")
// }

// отображение информации о пользователе, а также любую ошибку
func (pd PageData) newPageData(userName, email, errMsg string) PageData {
	return PageData{
		Username:  userName,
		Email:     email,
		UserError: UserError{ErrorMessage: errMsg},
	}
}

// выгрузка данных для страницы
type PageData struct {
	Username  string
	Email     string
	UserError UserError
}




/* ПРОТЕСТИРУЙ и потом сохрани сперва, перед рефакторингом.

	сноси нахер все поля в пейлд дейт и ставь User и поле для ошибок
	далее уже переделаешь сам метод newPageData и подкоректируешь 
	функ validateAndRender. Напиши копию заранее и все, потом удалишь старую

	что до структур с логин.го то там в User закидывай структ Персоналок.
	Далее у тебя будят универсальная структура и метод для регистра.го и логина.го
*/






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

// Обработчик для отображения HTML-страницы регистрации
func ShowRegistrationPage(w http.ResponseWriter, r *http.Request) {
	// link := "web/html/register.html"
	// http.ServeFile(w, r, link)

	utils.RenderTemplate(w, PageData{},
		// utils.RenderTemplate(w, PageData{UserError: UserError{}},
		"web/html/register.html",
		"web/html/navigation.html")
}
