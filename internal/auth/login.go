package auth

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"First_internet_store/internal/database"
	"First_internet_store/internal/utils"
)

// Deprecate???

// Функция для обработки запросов на авторизацию пользователя
func loginHandler_000________000(w http.ResponseWriter, r *http.Request) {
	// Проверяем, был ли отправлен POST-запрос
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Загружаем HTML-страницу для входа на сайт
	// loginPage, err := ioutil.ReadFile("templates/login.html") // Замените "path/to/login.html" на путь к вашему HTML-файлу
	// if err != nil {
	// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
	// 	log.Println("Error reading login page:", err)
	// 	return
	// }

	// -------------------------------------------------------------

	// if r.Method != http.MethodPost {
	// 	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	// 	return
	// }
	// Отправляем HTML-страницу входа на сайт в ответ на запрос
	http.ServeFile(w, r, "templates/login.html")
	// return // Добавляем return для завершения выполнения обработчика

	// Отправляем HTML-страницу в ответ на запрос
	// w.Header().Set("Content-Type", "text/html")
	// w.Write(loginPage)

	// -------------------------------------------------------------

	// Подключаемся к базе данных PostgreSQL
	db, err := sql.Open("postgres", "user=postgres password=qwerty dbname=online_store sslmode=disable")
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		log.Fatal("Error connecting to the database:", err)
		return
	}
	defer db.Close()

	// Получаем данные, отправленные пользователем с формы входа на сайт
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
		log.Println("Error parsing form data:", err)
		return
	}

	// Извлекаем данные из формы
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	// Проверяем данные пользователя в базе данных
	var dbPassword string
	err = db.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&dbPassword)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		log.Println("Invalid username or password:", err)
		return
	}

	// Проверяем соответствие пароля
	if password != dbPassword {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		log.Println("Invalid username or password")
		return
	}

	// Выводим сообщение об успешной авторизации
	fmt.Fprintf(w, "Welcome, %s!", username)
}

// ==========================================================================================
// ==========================================================================================

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/home/mrx/Documents/Programm Go/Results/2024.04.19_First_internet_store/First_internet_store/web/views/login.html")
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Подключаемся к базе данных PostgreSQL
	// db, err := sql.Open("postgres", "user=postgres password=qwerty dbname=online_store sslmode=disable")
	// if err != nil {
	// 	http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
	// 	log.Fatal("Error connecting to the database:", err)
	// 	return
	// }
	// defer db.Close()

	db, err := database.Connect()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		log.Fatal("Error connecting to the database:", err)
		return
	}
	defer db.Close()

	// Получаем данные, отправленные пользователем с формы входа
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
		log.Fatal("Error parsing form data:", err)
		return
	}

	// Извлекаем данные из формы
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	// Запрос к базе данных для проверки учётных данных
	var storedPassword string
	err = db.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&storedPassword)
	if err != nil {
		// В случае ошибки либо пользователя не существует, либо произошла ошибка в запросе
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Проверяем пароль
	if storedPassword != password {
		// Пароль не совпадает
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Если всё успешно, можно выполнить действия после успешной авторизации, например, установить сессию
	session, err := utils.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Здесь можно установить какие-то данные в сессию, например, идентификатор пользователя
	session.Values["user_id"] = username
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Перенаправляем пользователя на какую-то страницу после успешной авторизации
	http.Redirect(w, r, "/user-dashboard", http.StatusSeeOther)
}

// Обработчик для страницы с панелью управления пользователем после успешной авторизации
func UserDashboardHandler(w http.ResponseWriter, r *http.Request) {

	session, err := utils.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверяем, авторизован ли пользователь
	if session.Values["user_id"] == nil {
		// Если пользователь не авторизован, перенаправляем на страницу входа
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Получаем имя пользователя из сессии
	username := session.Values["user_id"].(string)

	// Загружаем HTML-шаблон страницы панели управления пользователя
	link := "/home/mrx/Documents/Programm Go/Results/2024.04.19_First_internet_store/First_internet_store/web/views/user_dashboard.html"
	tmpl := template.Must(template.ParseFiles(link))

	// Передаем имя пользователя в HTML-шаблон и отправляем его клиенту
	err = tmpl.Execute(w, map[string]interface{}{
		"Username": username,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Handling user dashboard request for:", username)
}
