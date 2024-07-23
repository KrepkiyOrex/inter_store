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

/*
	Здесь выполняю 2 разных запроса по отдельности т.к. если вместе сделать через
	JOIN, то вылазит ошибка на пустые данные БД и страницу на отображает в браузере.
	person_details изначально же с пустыми данными т.к. после регистрации пользователи
	обычно не заполняют сразу данные. ХЗ как ошибку убрать.
*/
// user profile
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

	// Получаем ID пользователя из куки
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

	// Выполняем SQL запрос для получения данных пользователя из таблицы users
	var user User
	err = db.QueryRow("SELECT username, email, user_id FROM users WHERE user_id = $1", userID).Scan(
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

	// Выполняем SQL запрос для получения данных из таблицы person_details
	var personDetails PersonDetails
	err = db.QueryRow("SELECT first_name, last_name, middle_name, phone FROM person_details WHERE user_id = $1", userID).Scan(
		&personDetails.First_name,
		&personDetails.Last_name,
		&personDetails.Middle_name,
		&personDetails.Phone)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No person details found for userID: %v, UserName: %s", userID, user.UserName)
			personDetails = PersonDetails{}
		} else {
			log.Println("Error querying person details:", err)
			http.Error(w, "Error retrieving person details", http.StatusInternalServerError)
			return
		}
	}

	// Создаем структуру данных для передачи в шаблон
	data := struct {
		User
		PersonDetails
	}{
		User:          user,
		PersonDetails: personDetails,
	}

	// Рендерим шаблон с данными пользователя
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

func EditProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Получение данных пользователя для заполнения формы
		db, err := database.Connect()
		if err != nil {
			http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
			log.Println("Error connecting to the database")
			return
		}
		defer db.Close()

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

		var user User
		err = db.QueryRow("SELECT username, email, user_id FROM users WHERE user_id = $1", userID).Scan(
			&user.UserName,
			&user.Email,
			&user.User_ID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		var personDetails PersonDetails
		err = db.QueryRow("SELECT first_name, last_name, middle_name, phone FROM person_details WHERE user_id = $1", userID).Scan(
			&personDetails.First_name,
			&personDetails.Last_name,
			&personDetails.Middle_name,
			&personDetails.Phone)
		if err != nil {
			if err == sql.ErrNoRows {
				personDetails = PersonDetails{}
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		data := struct {
			User
			PersonDetails
		}{
			User:          user,
			PersonDetails: personDetails,
		}

		utils.RenderTemplate(w, data,
			"web/html/edit_profile.html",
			"web/html/navigation.html")

	} else if r.Method == http.MethodPost {
		// Обработка данных формы и обновление информации о пользователе
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
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
			return
		}
		defer db.Close()

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

		// Обновление данных пользователя
		_, err = db.Exec("UPDATE users SET username = $1, email = $2 WHERE user_id = $3",
			username, email, userID)
		if err != nil {
			http.Error(w, "Unable to update user data", http.StatusInternalServerError)
			return
		}

		// Проверка наличия записи в таблице person_details
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM person_details WHERE user_id = $1", userID).Scan(&count)
		if err != nil {
			http.Error(w, "Unable to check person details", http.StatusInternalServerError)
			return
		}

		if count == 0 {
			// Вставка новой записи, если ее нет
			_, err = db.Exec("INSERT INTO person_details (first_name, last_name, middle_name, phone, user_id) VALUES ($1, $2, $3, $4, $5)",
				firstName, lastName, middleName, phone, userID)
			if err != nil {
				http.Error(w, "Unable to insert personal details", http.StatusInternalServerError)
				return
			}
		} else {
			// Обновление существующей записи
			_, err = db.Exec("UPDATE person_details SET first_name = $1, last_name = $2, middle_name = $3, phone = $4 WHERE user_id = $5",
				firstName, lastName, middleName, phone, userID)
			if err != nil {
				http.Error(w, "Unable to update personal details", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/account", http.StatusSeeOther)
	}
}

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, PageData{},
		"web/html/login.html",
		"web/html/navigation.html")
}

// Обработчик для страницы входа
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// При GET запросе отображаем форму входа
	if r.Method == "GET" {
		utils.RenderTemplate(w, PageData{}, "web/html/login.html", "web/html/navigation.html")
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
		data := PageData{}.newPageData(userName, "", "Invalid username or password")

		// data := PageData{
		// 	Username: userName,
		// 	UserError: UserError{ErrorMessage: "Invalid username or password"},
		// }

		utils.RenderTemplate(w, data, "web/html/login.html", "web/html/navigation.html")
		return
	}

	// устанавливаем ник в куки
	SetCookie(w, "userName", userName, time.Now().Add(24*time.Hour))

	// устанавливаем ID пользователя в куки
	SetCookie(w, "userID", strconv.Itoa(userID), time.Now().Add(24*time.Hour))

	// Перенаправляем на ЛК страницу
	http.Redirect(w, r, "/account", http.StatusFound)
}

// 123
type UserError struct {
	ErrorMessage string
}

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
	if claims["name"] != expectedUser.UserName {
		return fmt.Errorf("invalid name claim")
	}

	if claims["password"] != expectedUser.Password {
		return fmt.Errorf("invalid password claim")
	}

	return nil
}

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
