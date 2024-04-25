package api

import (
	"database/sql"
	"log"
	"net/http"

	"First_internet_store/internal/auth"
	"First_internet_store/internal/models"
	"First_internet_store/internal/others"

	"github.com/gorilla/mux"
)

type DB struct {
	*sql.DB
}

type Store struct {
	DB *DB
}

// Обработчики HTTP
func SetupRoutes() *mux.Router {
	// Создаем маршрутизатор
	router := mux.NewRouter()

	// Создаем экземпляр DB, используя функцию Connect() из пакета database
	// db, err := database.Connect()
	// if err != nil {
	// 	log.Fatal("Ошибка при подключении к базе данных:", err)
	// }

	// Создаем экземпляр Store, который содержит соединение с базой данных
	// store := &models.Store{DB: db}

	// connect := &models.DB{DB: store.db}

	// connect := &models.DB{DB: db}

	router.HandleFunc("/", others.GreetHandler)
	router.HandleFunc("/hello", others.HelloHandler)
	router.HandleFunc("/headers", others.HeadersHandler)

	// router.HandleFunc("/products", connect.ProductsHandler) // depr ??
	router.HandleFunc("/products", models.ProductsHandler)
	router.HandleFunc("/add-to-cart", models.AddToCartHandler)
	router.HandleFunc("/cart", models.ViewCartHandler)
	router.HandleFunc("/users-orders", models.UserOrdersHandler) // error "driver"

	// Обработчик для отображения страницы регистрации (GET)
	router.HandleFunc("/registration", auth.ShowRegistrationPage)
	router.HandleFunc("/register", auth.RegisterHandler) // Обработчик для страницы регистрации

	// Страница входа и её обработчик
	router.HandleFunc("/login", auth.LoginPageHandler).Methods("GET")
	router.HandleFunc("/login", auth.LoginHandler).Methods("POST")
	router.HandleFunc("/user-dashboard", auth.UserDashboardHandler) // Страница панели управления пользователя

	return router
}

// Launcher server
func StartServer() {
	router := SetupRoutes()

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("Listen and Server:", err)
	}
}
