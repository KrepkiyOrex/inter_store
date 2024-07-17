package api

import (
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"First_internet_store/internal/admin"
	"First_internet_store/internal/auth"
	"First_internet_store/internal/models"
	"First_internet_store/internal/others"

	"github.com/gorilla/mux"
)

// fs := http.FileServer(http.Dir("./css/")) // "static" - без этого НЕ пашет CSS! F***!
// http.Handle("/css/", http.StripPrefix("/css/", fs))

// http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))))

// fileServer := http.FileServer(http.Dir("./web/static/"))
// router.Handle("/static/", http.StripPrefix("/static", fileServer))

// router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

// Обработчики HTTP
func SetupRoutes() *mux.Router {
	// Создаем маршрутизатор
	router := mux.NewRouter()

	router.HandleFunc("/", models.ProductsHandler)
	// router.HandleFunc("/", others.GreetHandler)
	// router.HandleFunc("/hello", others.HelloHandler)
	// router.HandleFunc("/hello", models.HelloHandler)
	router.HandleFunc("/headers", others.HeadersHandler)

	// Настройка обработчика для статических файлов
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(".", "web", "static", strings.TrimPrefix(r.URL.Path, "/static/"))
		ext := filepath.Ext(filePath)
		mimeType := mime.TypeByExtension(ext)
		if mimeType != "" {
			w.Header().Set("Content-Type", mimeType)
		}
		http.ServeFile(w, r, filePath)
	})))

	// router.HandleFunc("/products", models.ProductsHandler)
	router.HandleFunc("/list", models.ListHandler)
	// router.HandleFunc("/add-to-cart", models.AddToCartHandler)
	// router.HandleFunc("/add-to-cart", models.AddToCartHandler).Methods("POST")
	router.HandleFunc("/users-orders", models.UserOrdersHandler) // error "driver"
	
	
	router.HandleFunc("/cart", models.ViewCartHandler)
	router.HandleFunc("/edit", models.EditProduct)

	// Обработчик для отображения страницы регистрации (GET)
	router.HandleFunc("/registration", auth.ShowRegistrationPage)
	router.HandleFunc("/register", auth.RegisterHandler) // Обработчик для страницы регистрации

	// Страница входа и её обработчики
	router.HandleFunc("/login", auth.LoginPageHandler).Methods("GET")
	router.HandleFunc("/login", auth.LoginHandler).Methods("POST")

	router.HandleFunc("/logout", auth.LogoutHandler) // Exit

	// deprecated из-за ненадобности
	// router.HandleFunc("/user-dashboard", auth.UserDashboardHandler) // Страница панели управления пользователя
	router.HandleFunc("/account", auth.Account) // profile

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
