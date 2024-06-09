package auth

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"First_internet_store/internal/database"

	"github.com/dgrijalva/jwt-go"
)

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	// http.ServeFile(w, r, "web/html/login.html")
	renderTemplate(w, UserCookie{}, "web/html/login.html", "web/html/navigation.html")
}

// Обработчик для страницы входа
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// При GET запросе отображаем форму входа
	if r.Method == "GET" {
		// utils.RenderTemplate(w, utils.UserCookie{}, "web/html/login.html", "web/html/navigation.html")
		renderTemplate(w, UserCookie{}, "web/html/login.html", "web/html/navigation.html")
		// return
	}

	// При POST запросе обрабатываем вход пользователя
	userName := r.FormValue("userName")
	password := r.FormValue("password")

	// Аутентифицируем пользователя
	_, err := authenticateUser(userName, password)
	if err != nil {
		// Если аутентификация не удалась, показываем форму входа снова с ошибкой
		date := UserCookie{
			ErrorMessage: "Invalid username or password",
		}

		renderTemplate(w, date, "web/html/login.html", "web/html/navigation.html")
		return
	}

	// Устанавливаем куку с именем пользователя
	http.SetCookie(w, &http.Cookie{
		Name:  "userName",
		Value: userName,
		Path:  "/",
	})
	// Перенаправляем на ЛК страницу
	http.Redirect(w, r, "/account", http.StatusFound)
}

type UserCookie struct {
	UserName     string
	ErrorMessage string
}

func renderTemplate(w http.ResponseWriter, data UserCookie, tmpl ...string) {
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

// Обработчик для выхода из аккаунта
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Удаляем куку с именем пользователя
	http.SetCookie(w, &http.Cookie{
		Name:   "userName",
		Value:  "",
		Path:   "/",
		MaxAge: -1, // Устанавливаем отрицательное время жизни, чтобы кука удалилась
	})

	// Перенаправляем на главную страницу
	http.Redirect(w, r, "/", http.StatusFound)
}

// DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED?
// DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED?
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

// DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED?
// DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED?
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

// DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED?
// DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED?
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
