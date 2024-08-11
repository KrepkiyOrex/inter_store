package utils

import (
	"net/http"
	"strconv"
)

type Product struct {
	Name  string
	Price int
	ID    int
}

type UserCookie struct {
	UserName  string
	Products  []Product
	UsersData UsersData // это поле явно линшенее тут!
}

type User struct {
	ID       int
	Username string
	Email    string
	Password string
}

type UsersData struct {
	Users []User
}

// Утильная функция для получения userID из куки
func GetUserIDFromCookie(r *http.Request) (int, error) {
	cookie, err := r.Cookie("userID")
	if err != nil || cookie.Value == "" {
		return 0, err
	}
	return strconv.Atoi(cookie.Value)
}
