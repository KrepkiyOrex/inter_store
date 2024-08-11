package main

import (
	"log"

	"First_internet_store/internal/api"
)

func main() {
	// Инициализируем хранилище сессий
	// utils.Store = sessions.NewCookieStore([]byte("your-secret-key"))

	// log.Println("Connecting to the database successfully")
	log.Println("Server starts")

	api.StartServer()
}
