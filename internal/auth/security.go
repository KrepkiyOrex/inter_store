package auth

import (
	"github.com/KrepkiyOrex/inter_store/internal/database"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func GetCookieUserID(w http.ResponseWriter, r *http.Request) (int, error) {
	cookie, err := r.Cookie("userID")
	if err != nil {
		if err == http.ErrNoCookie {
			return 0, fmt.Errorf("user ID cookie not found")
		}
		return 0, err
	}

	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID")
	}

	return userID, nil
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

// для проверки конкретных утверждений
func validateClaims(claims jwt.MapClaims, expectedUser User) error {
	if claims["name"] != expectedUser.UserName {
		return fmt.Errorf("invalid name claim")
	}

	if claims["password"] != expectedUser.Password {
		return fmt.Errorf("invalid password claim")
	}

	return nil
}

// аутентификация пользователя и получения его идентификатора
func authenticateUser(username, password string) (int, error) {
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Error connecting to the database", err)
	}
	defer db.Close()

	var userID int
	var storedUsername, storedPassword string

	// Выполняем запрос к базе данных для проверки логина и пароля
	err = db.QueryRow(`
			SELECT user_id, username, password 
				FROM users 
				WHERE username = $1`,
		username).Scan(&userID, &storedUsername, &storedPassword)

	if err != nil {
		// В случае ошибки или неверных учетных данных возвращаем ошибку аутентификации
		return 0, err
	}

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
