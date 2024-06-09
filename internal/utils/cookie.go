package utils

type Product struct {
	Name  string
	Price int
	ID    int
}

type UserCookie struct {
	UserName string
	Products []Product
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