package auth

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func TestExtractToken(t *testing.T) {
	// Создаем запрос с заголовком Authorization содержащим токен
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer myToken123")

	// Извлекаем токен из запроса
	token := ExtractToken(req)

	// Проверяем, что токен извлечен корректно
	expectedToken := "myToken123"
	if token != expectedToken {
		t.Errorf("Extracted token is incorrect, expected %s, got %s", expectedToken, token)
	}

	// Создаем запрос без заголовка Authorization
	reqWithoutHeader := httptest.NewRequest("GET", "/", nil)

	// Извлекаем токен из запроса без заголовка Authorization
	tokenFromEmptyRequest := ExtractToken(reqWithoutHeader)

	// Проверяем, что функция вернула пустую строку, так как заголовок Authorization отсутствует
	if tokenFromEmptyRequest != "" {
		t.Error("Extracted token from empty request should be an empty string")
	}
}

func TestCreateToken(t *testing.T) {
	userID := 123 // Пример идентификатора пользователя

	// Вызываем функцию createToken для создания токена
	token, err := createToken(userID)

	// Проверяем, что ошибки нет
	if err != nil {
		t.Errorf("createToken returned an error: %v", err)
	}

	// Декодируем токен, чтобы проверить его содержимое
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		t.Errorf("failed to parse token: %v", err)
	}

	// Проверяем, что токен содержит правильный идентификатор пользователя
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Error("token claims are not of type MapClaims")
	}

	userIDClaim, ok := claims["user_id"].(float64) // Тип float64 потому что MapClaims хранит все данные в интерфейсах
	if !ok {
		t.Error("user_id claim is not of type float64")
	}

	if int(userIDClaim) != userID {
		t.Errorf("token contains incorrect user_id, expected %d, got %d", userID, int(userIDClaim))
	}

	// Проверяем, что время истечения срока действия токена установлено корректно
	expClaim, ok := claims["exp"].(float64)
	if !ok {
		t.Error("exp claim is not of type float64")
	}

	expTime := time.Unix(int64(expClaim), 0)
	expectedExpTime := time.Now().Add(time.Hour * 24) // Токен действителен 24 часа
	if expTime.Sub(expectedExpTime) > time.Second {
		t.Errorf("token expiration time is incorrect, expected %v, got %v", expectedExpTime, expTime)
	}
}

// ================================================
