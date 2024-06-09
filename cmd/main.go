package main

import (
	"log"

	"github.com/gorilla/sessions"

	"First_internet_store/internal/api"
	"First_internet_store/internal/utils"
)

func main() {
	// Инициализируем хранилище сессий
	utils.Store = sessions.NewCookieStore([]byte("your-secret-key"))

	// log.Println("Connecting to the database successfully")
	log.Println("Server starts")

	api.StartServer()
}
