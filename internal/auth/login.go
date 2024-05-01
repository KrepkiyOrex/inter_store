package auth

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"First_internet_store/internal/database"
	"First_internet_store/internal/utils"

	"github.com/dgrijalva/jwt-go"
)

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/views/login.html")
}

// Обработчик для аутентификации пользователя и выдачи токена
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем данные, отправленные пользователем с формы аутентификации
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
		log.Println("Error parsing form data:", err)
		return
	}

	// Извлекаем данные из формы
	userName := r.Form.Get("username")
	password := r.Form.Get("password")

	// Проверяем данные пользователя в базе данных
	// V Здесь должна быть проверка пароля и другая бизнес-логика аутентификации
	// Если аутентификация проходит успешно, создаем токен для пользователя

	// Предположим, что у вас есть функция для проверки пользовательских данных и получения ID пользователя из базы данных
	userID, err := authenticateUser(userName, password)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		log.Println("Authentication failed:", err)
		return
	}

	// Создаем токен для пользователя
	token, err := createToken(userID)
	if err != nil {
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		log.Println("Failed to create token:", err)
		return
	}

	// // Отправляем токен в ответ на запрос
	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(map[string]string{"token": token})

	// Устанавливаем токен в HTTP cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		// Другие атрибуты, такие как Secure и SameSite, могут быть также установлены в соответствии с вашими требованиями безопасности.
	})

	// (получает сессию из хранилища. Если сессия с указанным именем не существует, она будет создана.)
	// Если всё успешно, можно выполнить действия после успешной авторизации, например, установить сессию
	session, err := utils.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// (устанавливает значение "value" для ключа "key" в данных сессии. Можно
	// установить любые данные, для сохранения в сессию.)
	// записиваем данные клиента в сессию
	session.Values["user_id"] = userName
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ==============================================================

	// Редирект на другую страницу после успешной аутентификации
	http.Redirect(w, r, "/user-dashboard", http.StatusFound)
}

// Функция для аутентификации пользователя и получения его идентификатора
func authenticateUser(username, password string) (int, error) {
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Error conecting to the database", err)
	}
	defer db.Close()

	var userID int
	var storedUsername, storedPassword string

	// Выполняем запрос к базе данных для проверки логина и пароля
	err = db.QueryRow("SELECT id, username, password FROM users WHERE username = $1",
		username).Scan(&userID, &storedUsername, &storedPassword)

	if err != nil {
		// В случае ошибки или неверных учетных данных возвращаем ошибку аутентификации
		return 0, err
	}

	// Проверяем соответствие имени пользователя и хэшированного пароля из базы данных с предоставленными данными
	// Здесь должна быть ваша логика хэширования и проверки пароля
	if storedUsername != username || storedPassword != password {
		return 0, errors.New("Invalid username or password")
	}

	// Возвращаем идентификатор пользователя, если аутентификация прошла успешно
	return userID, nil
}

// Функция для создания JWT токена на основе идентификатора пользователя
func createToken(userID int) (string, error) {
	// Задаем секретный ключ для подписи токена (он должен быть безопасно храниться и не раскрываться)
	secretKey := []byte("my_secret_key")

	// Создаем новый JWT токен
	token := jwt.New(jwt.SigningMethodHS256)

	// Создаем claims для токена, например, указываем идентификатор пользователя в нем
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	// Добавляем дополнительные поля в claims, если это необходимо
	// Например: claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	// Устанавливаем время истечения срока действия токена
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Например, токен действителен 24 часа

	// Подписываем токен с использованием секретного ключа
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Функция для извлечения токена из заголовка запроса
func ExtractToken(r *http.Request) string {
	// Получаем значение заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	// Проверяем, что заголовок не пустой и начинается с "Bearer "
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		// Извлекаем токен, убирая "Bearer " из начала строки
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}

type User struct {
	ID    int
	Name  string
	Email string
}

// Функция для проверки токена и извлечения информации о пользователе
func GetUserFromToken(tokenString string) (User, error) {
	// Установка секретного ключа для проверки подписи токена
	secretKey := []byte("my_secret_key")

	// Парсим токен из строки, используя секретный ключ
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Возвращаем установленный секретный ключ для проверки подписи токена
		return secretKey, nil
	})

	// Обработка ошибок при парсинге токена
	if err != nil {
		return User{}, err
	}

	// Проверяем, что токен валиден
	if !token.Valid {
		return User{}, errors.New("Invalid token")
	}

	// Извлекаем данные о пользователе из токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return User{}, errors.New("Failed to parse claims")
	}

	// Получаем необходимые данные о пользователе из токена
	userID, ok := claims["user_id"].(int)
	if !ok {
		return User{}, errors.New("Failed to parse user ID")
	}

	// Возвращаем данные о пользователе
	user := User{
		ID: userID,
		// Дополнительные данные о пользователе, которые могут быть в токене
		Name:  claims["name"].(string),
		Email: claims["email"].(string),
	}

	return user, nil
}

// информация о пользователе (user), которую можно использовать
// для выполнения нужных действий в обработчике (на будущее пока оставил)
// func YourHandler(w http.ResponseWriter, r *http.Request) {
// 	// Извлекаем токен из заголовка запроса
// 	tokenString := ExtractToken(r)
// 	if tokenString == "" {
// 		http.Error(w, "No token provided", http.StatusUnauthorized)
// 		return
// 	}

// 	// Получаем информацию о пользователе из токена
// 	user, err := GetUserFromToken(tokenString)
// 	if err != nil {
// 		http.Error(w, "Failed to authenticate user", http.StatusUnauthorized)
// 		return
// 	}

// 	// Теперь у вас есть информация о пользователе (user), которую вы можете использовать
// 	// для выполнения нужных действий в вашем обработчике
// }

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
	// username, ok  := session.Values["user_id"].(string)
	// if !ok {
	// 	// Если user_id не является строкой, обработайте эту ситуацию
	// 	// Например, можно сделать перенаправление на страницу входа или отправить ошибку
	// 	http.Error(w, "Failed to retrieve username from session", http.StatusInternalServerError)
	// 	return
	// }

	// Загружаем HTML-шаблон страницы панели управления пользователя
	link := "web/views/user_dashboard.html"
	tmpl := template.Must(template.ParseFiles(link))

	// Передаем имя пользователя в HTML-шаблон и отправляем его клиенту
	err = tmpl.Execute(w, map[string]interface{}{
		"Username": username,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// log.Println("Handling username:", username)

	log.Println("Handling user dashboard request for:", username) // work
}

// func LoginHandler(w http.ResponseWriter, r *http.Request) {
// 	db, err := database.Connect()
// 	if err != nil {
// 		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
// 		log.Fatal("Error connecting to the database:", err)
// 		return
// 	}
// 	defer db.Close()

// 	// Получаем данные, отправленные пользователем с формы входа
// 	err = r.ParseForm()
// 	if err != nil {
// 		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
// 		log.Fatal("Error parsing form data:", err)
// 		return
// 	}

// 	// Извлекаем данные из формы
// 	username := r.Form.Get("username")
// 	password := r.Form.Get("password")

// // Запрос к базе данных для проверки учётных данных
// var storedPassword string
// err = db.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&storedPassword)
// if err != nil {
// 	// В случае ошибки либо пользователя не существует, либо произошла ошибка в запросе
// 	http.Error(w, "Invalid username or password", http.StatusUnauthorized)
// 	return
// }

// // Проверяем пароль
// if storedPassword != password {
// 	// Пароль не совпадает
// 	http.Error(w, "Invalid username or password", http.StatusUnauthorized)
// 	return
// }

// 	// Если всё успешно, можно выполнить действия после успешной авторизации, например, установить сессию
// 	session, err := utils.Store.Get(r, "session-name")
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Здесь можно установить какие-то данные в сессию, например, идентификатор пользователя
// 	session.Values["user_id"] = username
// 	err = session.Save(r, w)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Перенаправляем пользователя на какую-то страницу после успешной авторизации
// 	http.Redirect(w, r, "/user-dashboard", http.StatusSeeOther)
// }

// =================================================================
// =================================================================

// Функция для обработки запросов на авторизацию пользователя
// func loginHandler_000________000(w http.ResponseWriter, r *http.Request) {
// 	// Проверяем, был ли отправлен POST-запрос
// 	if r.Method != http.MethodPost && r.Method != http.MethodGet {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Загружаем HTML-страницу для входа на сайт
// 	// loginPage, err := ioutil.ReadFile("templates/login.html") // Замените "path/to/login.html" на путь к вашему HTML-файлу
// 	// if err != nil {
// 	// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
// 	// 	log.Println("Error reading login page:", err)
// 	// 	return
// 	// }

// 	// -------------------------------------------------------------

// 	// if r.Method != http.MethodPost {
// 	// 	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 	// 	return
// 	// }
// 	// Отправляем HTML-страницу входа на сайт в ответ на запрос
// 	http.ServeFile(w, r, "templates/login.html")
// 	// return // Добавляем return для завершения выполнения обработчика

// 	// Отправляем HTML-страницу в ответ на запрос
// 	// w.Header().Set("Content-Type", "text/html")
// 	// w.Write(loginPage)

// 	// -------------------------------------------------------------

// 	// Подключаемся к базе данных PostgreSQL
// 	db, err := sql.Open("postgres", "user=postgres password=qwerty dbname=online_store sslmode=disable")
// 	if err != nil {
// 		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
// 		log.Fatal("Error connecting to the database:", err)
// 		return
// 	}
// 	defer db.Close()

// 	// Получаем данные, отправленные пользователем с формы входа на сайт
// 	err = r.ParseForm()
// 	if err != nil {
// 		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
// 		log.Println("Error parsing form data:", err)
// 		return
// 	}

// 	// Извлекаем данные из формы
// 	username := r.Form.Get("username")
// 	password := r.Form.Get("password")

// 	// Проверяем данные пользователя в базе данных
// 	var dbPassword string
// 	err = db.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&dbPassword)
// 	if err != nil {
// 		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
// 		log.Println("Invalid username or password:", err)
// 		return
// 	}

// 	// Проверяем соответствие пароля
// 	if password != dbPassword {
// 		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
// 		log.Println("Invalid username or password")
// 		return
// 	}

// 	// Выводим сообщение об успешной авторизации
// 	fmt.Fprintf(w, "Welcome, %s!", username)
// }
