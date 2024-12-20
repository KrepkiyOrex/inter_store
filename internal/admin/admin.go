package admin

import (
	"github.com/KrepkiyOrex/inter_store/internal/auth"
	"github.com/KrepkiyOrex/inter_store/internal/database"
	"github.com/KrepkiyOrex/inter_store/internal/utils"
	"net/http"
	
	log "github.com/sirupsen/logrus"
	"github.com/gorilla/mux"
)

type User struct {
	ID       int
	Username string
	Password string
	Email    string
}

type PageData struct {
	UsersData
	UserCookie
}

type UsersData struct {
	Users []User
}

type UserCookie struct {
	UserName string
}

// удаление юзера из списка БД
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		log.Warn("Invalid request method for deleting user")
		return
	}

	// Получение ID пользователя из URL
	vars := mux.Vars(r)
	userID := vars["id"]
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		log.Warn("Missing user ID in request")
		return
	}

	// Подключение к базе данных
	db, err := database.Connect()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		log.Error("Error connecting to the database:", err)
		return
	}
	defer db.Close()

	// Выполнение SQL-запроса для удаления пользователя
	query := "DELETE FROM users WHERE user_id = $1"
	log.Println("Executing query:", query, "with userID:", userID)
	_, err = db.Exec(query, userID)
	if err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		log.Error("Error deleting user:", err)
		return
	}

	// успешное удаление
	w.WriteHeader(http.StatusOK)
}

// Admin dashboard
func AdminPanel(w http.ResponseWriter, r *http.Request) {
	// Connect to the database
	db, err := database.Connect()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		log.Error("Error connecting to the database")
		return
	}
	defer db.Close()

	// Выполнение SQL запроса для выдачи всех данных из таблицы users
	rows, err := db.Query("SELECT username, password, email, user_id FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Error querying database:", err)
		return
	}
	defer rows.Close()

	var users []User

	// Чтение результатов запроса
	for rows.Next() {
		var user User
		err := rows.Scan(&user.Username, &user.Password, &user.Email, &user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error("Error scanning row:", err)
			return
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Error reading rows:", err)
		return
	}

	log.Info("Successfully retrieved users from database, count:", len(users))

	userName, _ := auth.GetUserName(r)

	// отображение информации о пользователе, а также любую ошибку
	data := PageData{}.newPageData(users, userName)

	utils.RenderTemplate(w, data,
		"web/html/admin.html",
		"web/html/navigation.html")
}

// отображение информации о пользователе, а также любую ошибку
func (pd PageData) newPageData(users []User, userName string) PageData {
	return PageData{
		UsersData: UsersData{
			Users: users,
		},
		UserCookie: UserCookie{
			UserName: userName,
		},
	}
}

