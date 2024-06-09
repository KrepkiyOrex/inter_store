package admin

import (
	"First_internet_store/internal/database"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type User struct {
	ID       int
	Username string
	Email    string
	Password string
}

// type PageData struct {
//     Users []User
// }

func renderTemplate(w http.ResponseWriter, data interface{}, tmpl ...string) {
	template, err := template.ParseFiles(tmpl...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = template.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Получение ID пользователя из URL
	vars := mux.Vars(r)
	userID := vars["id"]
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	// Подключение к базе данных
	db, err := database.Connect()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		log.Println("Error connecting to the database:", err)
		return
	}
	defer db.Close()

	// Выполнение SQL-запроса для удаления пользователя
	query := "DELETE FROM users WHERE id = $1"
	log.Println("Executing query:", query, "with userID:", userID)
	_, err = db.Exec(query, userID)
	if err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		log.Println("Error deleting user:", err)
		return
	}

	// Успешное удаление
	w.WriteHeader(http.StatusOK)
}

func AdminPanel(w http.ResponseWriter, r *http.Request) {
	// Connect to the database
	db, err := database.Connect()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		log.Println("Error connecting to the database")
		return
	}
	defer db.Close()

	// id | username | email | password
	// Выполнение SQL запроса для выдачи всех данных из таблицы users
	rows, err := db.Query("SELECT username, password, email, id FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
			return
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// data := UsersData{
	// 	Users: users,
	// }
	
	var userName string

	cookie, err := r.Cookie("userName")
	if err == nil {
		userName = cookie.Value
	}

	data := PageData{
		UsersData: UsersData{
			Users: users,
		},
		UserCookie: UserCookie{
			UserName: userName,
		},
	}

	renderTemplate(w, data, 
		"web/html/admin.html",
		"web/html/navigation.html")
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