package models

import (
	"database/sql"
	"net/http"

	"github.com/KrepkiyOrex/inter_store/internal/auth"
	"github.com/KrepkiyOrex/inter_store/internal/database"
	"github.com/KrepkiyOrex/inter_store/internal/utils"
	log "github.com/sirupsen/logrus"
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

// user profile
func Account(w http.ResponseWriter, r *http.Request) {
	// Подключаемся к базе данных
	db, err := database.Connect()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		log.Error("Error connecting to the database: ", err)
		return
	}
	defer db.Close()

	userID, err := auth.GetCookieUserID(w, r)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
	log.WithFields(log.Fields{
		"error": err,
		"route": "Account",
		"userIP": r.RemoteAddr,
	}).Error("Invalid user ID from cookie")
	return
	}
	log.Info("Parsed User ID: ", userID)

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
			log.Info("No user found for userID: ", userID)
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Error("Error querying user: ", err)
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
			log.Infof("No person details found for userID: %v, UserName: %s", userID, user.UserName)
			user.PersonDetails = PersonDetails{}
		} else {
			log.Println("Error querying person details:", err)
			http.Error(w, "Error retrieving person details", http.StatusInternalServerError)
			return
		}
	}

	userName, _ := auth.GetUserName(r)

	// создаем структуру данных для передачи в шаблон
	data := PageData{}.newPageDataAcc(user, userName)

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
