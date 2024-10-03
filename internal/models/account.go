package models

import (
	"github.com/KrepkiyOrex/inter_store/internal/auth"
	"github.com/KrepkiyOrex/inter_store/internal/database"
	"github.com/KrepkiyOrex/inter_store/internal/utils"
	"database/sql"
	"log"
	"net/http"
	"strconv"
)

type User struct {
	User_ID       int
	UserName      string
	Password      string
	Email         string
	PersonDetails PersonDetails
}

type PersonDetails struct {
	User_id     int
	First_name  string
	Last_name   string
	Middle_name string
	Address     string
	Phone       string
}

/*
	Здесь выполняю 2 разных запроса по отдельности т.к. если вместе сделать через
	JOIN, то вылазит ошибка на пустые данные БД и страницу на отображает в браузере.
	person_details изначально же с пустыми данными т.к. после регистрации пользователи
	обычно не заполняют сразу данные. ХЗ как ошибку убрать.
*/

// user profile
func Account(w http.ResponseWriter, r *http.Request) {
	// Подключаемся к базе данных
	db, err := database.Connect()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		log.Println("Error connecting to the database:", err)
		return
	}
	defer db.Close()

	// получаем ID пользователя из куки
	cookieID, err := r.Cookie("userID")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(w, "User not authenticated", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Error retrieving userID", http.StatusInternalServerError)
		return
	}

	userID, err := strconv.Atoi(cookieID.Value)
	if err != nil {
		http.Error(w, "Invalid userID format", http.StatusBadRequest)
		return
	}

	log.Println("Extracted userID from cookie:", userID)

	// запрос для получения данных пользователя из таблицы users
	userQuery := `SELECT username, email, user_id 
					FROM users 
					WHERE user_id = $1`
	var user User
	err = db.QueryRow(userQuery, userID).Scan(
		&user.UserName,
		&user.Email,
		&user.User_ID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No user found for userID:", userID)
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Println("Error querying user:", err)
			http.Error(w, "Error retrieving user data", http.StatusInternalServerError)
		}
		return
	}

	// запрос для получения персональных данных из таблицы person_details
	pdQuery := `SELECT first_name, last_name, middle_name, phone 
					FROM person_details 
					WHERE user_id = $1`
	err = db.QueryRow(pdQuery, userID).Scan(
		&user.PersonDetails.First_name,
		&user.PersonDetails.Last_name,
		&user.PersonDetails.Middle_name,
		&user.PersonDetails.Phone)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No person details found for userID: %v, UserName: %s", userID, user.UserName)
			user.PersonDetails = PersonDetails{}
		} else {
			log.Println("Error querying person details:", err)
			http.Error(w, "Error retrieving person details", http.StatusInternalServerError)
			return
		}
	}

	userName, _ := auth.GetUserName(r)

	// Создаем структуру данных для передачи в шаблон
	data := PageData{}.newPageDataAcc(user, userName)

	// Рендерим шаблон с данными пользователя
	utils.RenderTemplate(w, data,
		"web/html/account.html",
		"web/html/navigation.html")
}

// создание данных для страницы аккаунта
func (pd PageData) newPageDataAcc(user User, userName string) PageData {
	return PageData{
		User:       user,
		UserCookie: UserCookie{UserName: userName},
	}
}
