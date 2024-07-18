package auth

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"First_internet_store/internal/database"
	"First_internet_store/internal/utils"

	"github.com/dgrijalva/jwt-go"
)

// user profile
func Account(w http.ResponseWriter, r *http.Request) {
	// Извлекаем куку
	// cookie, err := r.Cookie("userName")
	// userName := "" // Значение по умолчанию, если кука не установлена
	// if err == nil {
	// 	userName = cookie.Value
	// }

	// -----------------------------------------------------------------

	// Connect to the database
	db, err := database.Connect()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		log.Println("Error connecting to the database")
		return
	}
	defer db.Close()

	// Получение ID пользователя из куки
	cookieID, err := r.Cookie("userID")
	if err != nil {
		http.Error(w, "Invalid userID", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(cookieID.Value)
	if err != nil {
		http.Error(w, "Invalid userID format", http.StatusBadRequest)
		return
	}

	log.Println("Extracted userID from cookie:", userID)

	// Выполнение SQL запроса для объединения данных из users и person_details по ID пользователя
	// var data struct {
	// 	User
	// 	PersonDetails
	// }

	// Выполнение SQL запроса для выдачи данных пользователя по его ID
	// var user User
	// err = db.QueryRow(
	// "SELECT username, password, email, user_id FROM users WHERE user_id = $1",
	// userID).Scan(&user.UserName, &user.Password, &user.Email, &user.ID)

	// Выполнение SQL запроса для выдачи данных пользователя по его ID
	// err = db.QueryRow(
	// 	"SELECT u.username, u.password, u.email, u.user_id, p.first_name, p.phone "+
	// 		"FROM users u "+
	// 		"JOIN person_details p ON u.user_id = p.user_id "+
	// 		"WHERE u.user_id = $1", userID).Scan(
	// 	&data.User.UserName,
	// 	&data.User.Password,
	// 	&data.User.Email,
	// 	&data.User.ID,
	// 	&data.PersonDetails.First_name,
	// 	&data.PersonDetails.Phone)

	// ===========================================================================
	
	/* 
		Здесь выполняю 2 разных запроса по отдельности т.к. если вместе сделать через
		JOIN, то вылазит ошибка на пустые данные БД и страницу на отображает в браузере.
		person_details изначально же с пустыми данными т.к. после регистрации обычно не 
		заполняют сразу данные. ХЗ как ошибку убрать 
	*/
	// Выполнение SQL запроса для получения данных пользователя из таблицы users
	var user User
	err = db.QueryRow("SELECT username, password, email, user_id FROM users WHERE user_id = $1", userID).Scan(
		&user.UserName,
		&user.Password,
		&user.Email,
		&user.User_ID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No user found for userID:", userID)
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Println("Query error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Выполнение SQL запроса для получения данных из таблицы person_details
	var personDetails PersonDetails
	err = db.QueryRow("SELECT first_name, last_name, middle_name, phone FROM person_details WHERE user_id = $1", userID).Scan(
		&personDetails.First_name,
		&personDetails.Last_name,
		&personDetails.Middle_name,
		&personDetails.Phone)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No person details found for userID:", userID)
			// Если данные не найдены, можно оставить personDetails пустым или установить значения по умолчанию
			personDetails = PersonDetails{}
		} else {
			log.Println("Query error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	log.Println("Executing query:", "with userID:", userID)

	// if err != nil {
	// 	if err == sql.ErrNoRows {
	// 		http.Error(w, "User not found", http.StatusNotFound)
	// 	} else {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	}
	// 	return
	// }

	// UserName, _ := GetUserName(r)

	// -----------------------------------------------------------------

	// data := UserCookie{
	// 	ID:       user.ID,
	// 	UserName: user.UserName,
	// 	Email:    user.Email,
	// }

	// Создание структуры данных для передачи в шаблон
	data := struct {
		User
		PersonDetails
	}{
		User:          user,
		PersonDetails: personDetails,
	}

	utils.RenderTemplate(w, data,
		"web/html/account.html",
		"web/html/navigation.html")
}

// Set userName and userID in cookie
func SetCookie(w http.ResponseWriter, name, value string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  expires,
	})
}

type User struct {
	User_ID  int
	UserName string
	Password string
	Email    string
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

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	// http.ServeFile(w, r, "web/html/login.html")
	utils.RenderTemplate(w, UserError{}, "web/html/login.html", "web/html/navigation.html")
}

// Обработчик для страницы входа
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// При GET запросе отображаем форму входа
	if r.Method == "GET" {
		utils.RenderTemplate(w, UserError{}, "web/html/login.html", "web/html/navigation.html")
		return
		// renderTemplate(w, UserCookie{}, "web/html/login.html", "web/html/navigation.html")
	}

	// При POST запросе обрабатываем вход пользователя
	userName := r.FormValue("userName")
	password := r.FormValue("password")

	// Аутентифицируем пользователя
	userID, err := authenticateUser(userName, password)
	if err != nil {
		// Если аутентификация не удалась, показываем форму входа снова с ошибкой
		date := UserError{
			ErrorMessage: "Invalid username or password",
		}

		utils.RenderTemplate(w, date, "web/html/login.html", "web/html/navigation.html")
		return
	}

	// Устанавливаем куку с именем пользователя
	// http.SetCookie(w, &http.Cookie{
	// 	Name:  "userName",
	// 	Value: userName,
	// 	Path:  "/",
	// })

	// ---------------------------------------------------

	// устанавливаем ник в куки
	SetCookie(w, "userName", userName, time.Now().Add(24*time.Hour))

	// устанавливаем ID пользователя в куки
	SetCookie(w, "userID", strconv.Itoa(userID), time.Now().Add(24*time.Hour))

	// Перенаправляем на ЛК страницу
	http.Redirect(w, r, "/account", http.StatusFound)
}

type UserError struct {
	// UserName     string
	ErrorMessage string
}

// func renderTemplate(w http.ResponseWriter, data interface{}, tmpl ...string) {
// 	template, err := template.ParseFiles(tmpl...)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	err = template.Execute(w, data)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }

// =====================================================================================================================

// Функция для создания JWT
func createJWT(user User, secretKey []byte) (string, error) {
	// Создаем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.User_ID,                          // Используем ID пользователя как subject
		"name":     user.UserName,                         // Используем имя пользователя
		"password": user.Password,                         // Используем пароль пользователя
		"email":    user.Email,                            // Используем электронную почту пользователя
		"exp":      time.Now().Add(time.Hour * 72).Unix(), // Устанавливаем срок действия токена на 72 часа
	})

	// Подписываем токен
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Функция для проверки JWT и конкретных утверждений
func validateJWT(tokenString string, secretKey []byte, expectedUser User) (bool, error) {
	// Парсим и проверяем токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return false, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Проверяем конкретные утверждения
		if err := validateClaims(claims, expectedUser); err != nil {
			return false, err
		}

		fmt.Println("Token valid, claims:", claims)
		return true, nil

	} else {
		return false, err
	}
}

// Функция для проверки конкретных утверждений
func validateClaims(claims jwt.MapClaims, expectedUser User) error {
	// if claims["sub"] != expectedUser.ID {
	//     return fmt.Errorf("invalid subject claim")
	// }

	if claims["name"] != expectedUser.UserName {
		return fmt.Errorf("invalid name claim")
	}

	if claims["password"] != expectedUser.Password {
		return fmt.Errorf("invalid password claim")
	}

	return nil
}

// =====================================================================================================================

// Функция для аутентификации пользователя и получения его идентификатора
func authenticateUser(username, password string) (int, error) {
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Error connecting to the database", err)
	}
	defer db.Close()

	var userID int
	var storedUsername, storedPassword string

	// Выполняем запрос к базе данных для проверки логина и пароля
	err = db.QueryRow("SELECT user_id, username, password FROM users WHERE username = $1",
		username).Scan(&userID, &storedUsername, &storedPassword)

	if err != nil {
		// В случае ошибки или неверных учетных данных возвращаем ошибку аутентификации
		return 0, err
	}

	// ------------------------------------------------------------------------------------------------------------------------

	secretKey := []byte("your-256-bit-secret") // Секретный ключ должен быть сильным
	// и случайным, а также защищен от доступа посторонних лиц.
	/* Однако важно хранить такие ключи в безопасном месте и не хранить их в открытом виде в
	исходном коде вашего приложения. Идеальным решением для хранения секретных ключей является
	использование переменных среды или других методов безопасного хранения конфиденциальной информации. */

	// сохраняем введенные данные пользователя при авторизации
	userFormValue := User{
		UserName: username,
		Password: password,
	}

	// Создаем JWT
	token, err := createJWT(userFormValue, secretKey)
	if err != nil {
		fmt.Println("Error creating token:", err)
		return 0, err
	}
	fmt.Println("Generated JWT:", token)

	// сохраняем данные пользователя, выгруженные с БД
	userDB := User{
		UserName: storedUsername,
		Password: storedPassword,
	}

	// Проверяем JWT
	valid, err := validateJWT(token, secretKey, userDB)
	if err != nil {
		fmt.Println("Error validating token:", err)
		return 0, err
		// return 0, errors.New("Invalid username or password")
	}

	if valid {
		fmt.Println("Token is valid!")
	} else {
		fmt.Println("Token is invalid!")
	}

	// ------------------------------------------------------------------------------------------------------------------------

	// Проверяем соответствие имени пользователя и хэшированного пароля из базы данных с предоставленными данными
	// Здесь должна быть ваша логика хэширования и проверки пароля
	// if storedUsername != username || storedPassword != password {
	// 	return 0, errors.New("Invalid username or password")
	// }

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

// УДАЛИТЬ ДАННУЮ ЗАКОМЕНТИРОВУННУЮ ФУНКУИЮ, ТОЛЬКО ПОСЛЕ ТОГО, КАК
// СДЕЛАЕШЬ АЛЬТЕРНАТИВУ. ЧИТАЙ В ФАЙЛЕ АРХИТЕКТУРА на 92 строчке про это.
// DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED?
// DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED?
// Функция для создания JWT токена на основе идентификатора пользователя
// func createToken(userID int) (string, error) {
// 	// Задаем секретный ключ для подписи токена (он должен быть безопасно храниться и не раскрываться)
// 	secretKey := []byte("my_secret_key")

// 	// Создаем новый JWT токен
// 	token := jwt.New(jwt.SigningMethodHS256)

// 	// Создаем claims для токена, например, указываем идентификатор пользователя в нем
// 	claims := token.Claims.(jwt.MapClaims)
// 	claims["user_id"] = userID
// 	// Добавляем дополнительные поля в claims, если это необходимо
// 	// Например: claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

// 	// Устанавливаем время истечения срока действия токена
// 	claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Например, токен действителен 24 часа

// 	// Подписываем токен с использованием секретного ключа
// 	tokenString, err := token.SignedString(secretKey)
// 	if err != nil {
// 		return "", err
// 	}

// 	return tokenString, nil
// }

// УДАЛИТЬ ДАННУЮ ЗАКОМЕНТИРОВУННУЮ ФУНКУИЮ, ТОЛЬКО ПОСЛЕ ТОГО, КАК
// СДЕЛАЕШЬ АЛЬТЕРНАТИВУ. ЧИТАЙ В ФАЙЛЕ АРХИТЕКТУРА на 92 строчке про это.
// DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED?
// DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED?
// Функция для извлечения токена из заголовка запроса
// func ExtractToken(r *http.Request) string {
// 	// Получаем значение заголовка Authorization
// 	authHeader := r.Header.Get("Authorization")
// 	// Проверяем, что заголовок не пустой и начинается с "Bearer "
// 	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
// 		// Извлекаем токен, убирая "Bearer " из начала строки
// 		return strings.TrimPrefix(authHeader, "Bearer ")
// 	}
// 	return ""
// }

// УДАЛИТЬ ДАННУЮ ЗАКОМЕНТИРОВУННУЮ ФУНКУИЮ, ТОЛЬКО ПОСЛЕ ТОГО, КАК
// СДЕЛАЕШЬ АЛЬТЕРНАТИВУ. ЧИТАЙ В ФАЙЛЕ АРХИТЕКТУРА на 92 строчке про это.
// DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED?
// DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED? // DEPRECATED?
// Функция для проверки токена и извлечения информации о пользователе
// func GetUserFromToken(tokenString string) (User, error) {
// 	// Установка секретного ключа для проверки подписи токена
// 	secretKey := []byte("my_secret_key")

// 	// Парсим токен из строки, используя секретный ключ
// 	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 		// Возвращаем установленный секретный ключ для проверки подписи токена
// 		return secretKey, nil
// 	})

// 	// Обработка ошибок при парсинге токена
// 	if err != nil {
// 		return User{}, err
// 	}

// 	// Проверяем, что токен валиден
// 	if !token.Valid {
// 		return User{}, errors.New("Invalid token")
// 	}

// 	// Извлекаем данные о пользователе из токена
// 	claims, ok := token.Claims.(jwt.MapClaims)
// 	if !ok {
// 		return User{}, errors.New("Failed to parse claims")
// 	}

// 	// Получаем необходимые данные о пользователе из токена
// 	userID, ok := claims["user_id"].(int)
// 	if !ok {
// 		return User{}, errors.New("Failed to parse user ID")
// 	}

// 	// Возвращаем данные о пользователе
// 	user := User{
// 		ID: userID,
// 		// Дополнительные данные о пользователе, которые могут быть в токене
// 		Name:  claims["name"].(string),
// 		Email: claims["email"].(string),
// 	}

// 	return user, nil
// }

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
