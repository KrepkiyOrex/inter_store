package api

import (
	"log"
	"net/http"

	"First_internet_store/internal/admin"
	"First_internet_store/internal/auth"
	"First_internet_store/internal/models"
	"First_internet_store/internal/others"

	"github.com/gorilla/mux"
)

// Обработчики HTTP
func SetupRoutes() *mux.Router {
	// Создаем маршрутизатор
	router := mux.NewRouter()

	router.HandleFunc("/", others.GreetHandler)
	// router.HandleFunc("/hello", others.HelloHandler)
	router.HandleFunc("/hello", models.HelloHandler)
	router.HandleFunc("/headers", others.HeadersHandler)

	fs := http.FileServer(http.Dir("./web")) // "static" - без этого НЕ пашет CSS! F***!
	http.Handle("/web/", http.StripPrefix("/web/", fs))

	router.HandleFunc("/products", models.ProductsHandler)
	router.HandleFunc("/list", models.ListHandler)
	// router.HandleFunc("/add-to-cart", models.AddToCartHandler)
	router.HandleFunc("/add-to-cart", models.AddToCartHandler).Methods("POST")
	router.HandleFunc("/cart", models.ViewCartHandler)
	router.HandleFunc("/users-orders", models.UserOrdersHandler) // error "driver"

	// Обработчик для отображения страницы регистрации (GET)
	router.HandleFunc("/registration", auth.ShowRegistrationPage)
	router.HandleFunc("/register", auth.RegisterHandler) // Обработчик для страницы регистрации

	// Страница входа и её обработчик
	router.HandleFunc("/login", auth.LoginPageHandler).Methods("GET")

	// !!! ИСПРАВЬ косяк с методом "пост" auth.LoginHandler !!!
	router.HandleFunc("/login", auth.LoginHandler).Methods("POST")
	router.HandleFunc("/user-dashboard", auth.UserDashboardHandler) // Страница панели управления пользователя
	router.HandleFunc("/account", models.Account) // profile

	router.HandleFunc("/administrator", admin.AdminPanel) // admin panel
	router.HandleFunc("/administrator/{id}", admin.DeleteUser).Methods("DELETE")

	return router
}

// Launcher server
func StartServer() {
	router := SetupRoutes()

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("Listen and Server:", err)
	}

	/*
			Это более короткий способ?

			// Запускаем сервер
		    log.Fatal(http.ListenAndServe(":8080", nil))
	*/
}
