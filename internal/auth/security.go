package auth

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/KrepkiyOrex/inter_store/internal/database"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

func GetCookieUserID(w http.ResponseWriter, r *http.Request) (int, error) {
	cookie, err := r.Cookie("userID")
	if err != nil {
		if err == http.ErrNoCookie {
			log.Warn("User ID cookie not found")
			return 0, fmt.Errorf("user ID cookie not found")
		}
		log.Error("Error retrieving user ID cookie: ", err)
		return 0, err
	}

	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		log.Error("Invalid user ID in cookie, value: ", cookie.Value)
		return 0, fmt.Errorf("invalid user ID")
	}

	log.Info("Successfully retrieved User ID from cookie: ", userID)
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

// для создания JWT
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

// для проверки JWT и конкретных утверждений
func validateJWT(tokenString string, secretKey []byte, expectedUser User) (bool, error) {
	// Парсим и проверяем токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Errorf("unexpected signing method: %v", token.Header["alg"])
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

		log.Info("Token valid, claims: ", claims)
		return true, nil

	} else {
		return false, err
	}
}

// для проверки конкретных утверждений
func validateClaims(claims jwt.MapClaims, expectedUser User) error {
	if claims["name"] != expectedUser.UserName {
		log.Error("invalid name claim")
		return fmt.Errorf("invalid name claim")
	}

	if claims["password"] != expectedUser.Password {
		log.Error("invalid password claim")
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

	// secretKey := []byte("your-256-bit-secret") // Секретный ключ должен быть сильным
	// и случайным, а также защищен от доступа посторонних лиц.
	/* Однако важно хранить такие ключи в безопасном месте и не хранить их в открытом виде в
	исходном коде вашего приложения. Идеальным решением для хранения секретных ключей является
	использование переменных среды или других методов безопасного хранения конфиденциальной информации. */

	/*
		временная мера:
		export JWT_SECRET_KEY="mD$%k7L#jQ9*XYM6t@wJk"
		go run ./cmd .
	*/

	secretKey := os.Getenv("JWT_SECRET_KEY")
	log.Warn("[TOP SECRET!!!] Secret key: ", secretKey) // Выводим секретный ключ для проверки
	if secretKey == "" {
		log.Fatal("Secret key is not set in environment variables")
	}

	secretKeyBytes := []byte(secretKey)

	// сохраняем введенные данные пользователя при авторизации
	userFormValue := User{
		UserName: username,
		Password: password,
	}

	// создаем JWT
	token, err := createJWT(userFormValue, secretKeyBytes)
	if err != nil {
		log.Error("Error creating token: ", err)
		return 0, err
	}
	log.Info("Generated JWT: ", token)

	// сохраняем данные пользователя, выгруженные с БД
	userDB := User{
		UserName: storedUsername,
		Password: storedPassword,
	}

	// проверяем JWT
	valid, err := validateJWT(token, secretKeyBytes, userDB)
	if err != nil {
		log.Error("Error validating token: ", err)
		return 0, err
	}

	if valid {
		log.Info("Token is valid!")
	} else {
		log.Info("Token is invalid!")
	}

	// возвращаем идентификатор пользователя, если аутентификация прошла успешно
	return userID, nil
}
