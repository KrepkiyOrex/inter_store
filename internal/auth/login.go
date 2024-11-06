package auth

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/KrepkiyOrex/inter_store/internal/database"
	"github.com/KrepkiyOrex/inter_store/internal/utils"
	log "github.com/sirupsen/logrus"
)

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, PageData{},
		"web/html/login.html",
		"web/html/navigation.html")
}

// Обработчик для страницы входа
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// При GET запросе отображаем форму входа
	if r.Method == "GET" {
		utils.RenderTemplate(w, PageData{},
			"web/html/login.html",
			"web/html/navigation.html")
		return
	}

	// При POST запросе обрабатываем вход пользователя
	userName := r.FormValue("userName")
	password := r.FormValue("password")

	// Аутентифицируем пользователя
	userID, err := authenticateUser(userName, password)
	if err != nil {
		// Если аутентификация не удалась, показываем форму входа снова с ошибкой
		data := PageData{}.newPageData(userName, "", "Invalid username or password")

		utils.RenderTemplate(w, data,
			"web/html/login.html",
			"web/html/navigation.html")
		return
	}

	// устанавливаем ник в куки
	SetCookie(w, "userName", userName, time.Now().Add(24*time.Hour))

	// устанавливаем ID пользователя в куки
	SetCookie(w, "userID", strconv.Itoa(userID), time.Now().Add(24*time.Hour))

	// Перенаправляем на ЛК страницу
	http.Redirect(w, r, "/account", http.StatusFound)
}

// обработчик для выхода из аккаунта. Удаляет куки с информацией о пользователе
// и перенаправляет на главную страницу
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Удаляем куку с именем пользователя
	http.SetCookie(w, &http.Cookie{
		Name:   "userName",
		Value:  "",
		Path:   "/",
		MaxAge: -1, // Устанавливаем отрицательное время жизни, чтобы кука удалилась
	})

	// удаляем ID пользователя
	http.SetCookie(w, &http.Cookie{
		Name:   "userID",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Перенаправляем на главную страницу
	http.Redirect(w, r, "/", http.StatusFound)
}

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

type UserCookie struct {
	ID       int
	UserName string
	Password string
	Email    string
}

// редактор персональных данных пользователя в ЛК
func EditProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Получение данных пользователя для заполнения формы
		db, err := database.Connect()
		if err != nil {
			http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
			log.Error("Error connecting to the database: ", err)
			return
		}
		defer db.Close()

		userID, err := GetCookieUserID(w, r)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusInternalServerError)
			log.Error("Invalid user ID: ", err)
		}
		log.Info("Parsed User ID: ", userID)

		var user User
		queryUsers := `SELECT username, email, user_id 
						FROM users 
						WHERE user_id = $1`
		err = db.QueryRow(queryUsers, userID).Scan(
			&user.UserName,
			&user.Email,
			&user.User_ID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found", http.StatusNotFound)
				log.Warn("User not found for User ID: ", userID)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Error("Error querying user data: ", err)
			}
			return
		}

		queryPD := `SELECT first_name, last_name, middle_name, phone 
						FROM person_details 
						WHERE user_id = $1`
		err = db.QueryRow(queryPD, userID).Scan(
			&user.PersonDetails.First_name,
			&user.PersonDetails.Last_name,
			&user.PersonDetails.Middle_name,
			&user.PersonDetails.Phone)
		if err != nil {
			if err == sql.ErrNoRows {
				user.PersonDetails = PersonDetails{}
				log.Warn("No personal details found for User ID: ", userID)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Error("Error querying personal details: ", err)
				return
			}
		}

		data := PageData{User: user}

		utils.RenderTemplate(w, data,
			"web/html/edit_profile.html",
			"web/html/navigation.html")

	} else if r.Method == http.MethodPost {
		// Обработка данных формы и обновление информации о пользователе
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			log.Error("Error parsing form: ", err)
			return
		}

		username := r.FormValue("username")
		email := r.FormValue("email")
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		middleName := r.FormValue("middle_name")
		phone := r.FormValue("phone")

		db, err := database.Connect()
		if err != nil {
			http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
			log.Error("Error connecting to the database: ", err)
			return
		}
		defer db.Close()

		userID, err := GetCookieUserID(w, r)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusInternalServerError)
			log.Error("Invalid user ID: ", err)
		}
		log.Info("Parsed User ID: ", userID)

		// обновление данных пользователя
		_, err = db.Exec(`
				UPDATE users 
					SET username = $1, email = $2 
					WHERE user_id = $3`,
			username, email, userID)
		if err != nil {
			http.Error(w, "Unable to update user data", http.StatusInternalServerError)
			log.Error("Unable to update user data for User ID: ", userID, "Error: ", err)
			return
		}

		log.Info("Checking if personal details exist for User ID: ", userID)
		var count int
		err = db.QueryRow(`
				SELECT COUNT(*) 
					FROM person_details 
					WHERE user_id = $1`,
			userID).Scan(&count)
		if err != nil {
			http.Error(w, "Unable to check person details", http.StatusInternalServerError)
			log.Error("Unable to check person details for User ID: ", userID, "Error: ", err)
			return
		}

		if count == 0 {
			_, err = db.Exec(`
					INSERT INTO person_details 
						(first_name, last_name, middle_name, phone, user_id) 
						VALUES ($1, $2, $3, $4, $5)`,
				firstName, lastName, middleName, phone, userID)
			if err != nil {
				http.Error(w, "Unable to insert personal details", http.StatusInternalServerError)
				log.Error("Unable to insert personal details for User ID: ", userID, "Error: ", err)
				return
			}
		} else {
			_, err = db.Exec(`
					UPDATE person_details 
						SET first_name = $1, last_name = $2, middle_name = $3, phone = $4 
						WHERE user_id = $5`,
				firstName, lastName, middleName, phone, userID)
			if err != nil {
				http.Error(w, "Unable to update personal details", http.StatusInternalServerError)
				log.Error("Unable to update personal details for User ID: ", userID, "Error: ", err)
				return
			}
			log.Info("Updated personal details for User ID: ", userID)
		}

		http.Redirect(w, r, "/account", http.StatusSeeOther)
	}
}
