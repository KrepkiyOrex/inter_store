package api

import (
	"log"
	"net/http"

	"First_internet_store/internal/admin"
	"First_internet_store/internal/auth"
	"First_internet_store/internal/database"
	"First_internet_store/internal/models"
	"First_internet_store/internal/others"

	"github.com/gorilla/mux"
)

// AuthMiddleware проверяет, авторизован ли пользователь
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieID, err := r.Cookie("userID")
		if err != nil || cookieID.Value == "" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r) // передать управление следующему хендлеру
	})
}

// обработчики HTTP
func SetupRoutes() *mux.Router {
	// создаем маршрутизатор
	router := mux.NewRouter()

	database.InitRedis()      // initialization redis
	database.InitMongoClint() // initialization mongoDB

	router.HandleFunc("/item/{id:[0-9a-fA-F]{24}}", models.HandlerItemRequest) // используем маршрут с параметром id

	router.HandleFunc("/create-new-item", models.CreateNewItemHandler).Methods("POST")
	router.HandleFunc("/edit-item/{id:[0-9a-fA-F]{24}}", models.EditItemHandler).Methods("GET")
	router.HandleFunc("/update-item/{id:[0-9a-fA-F]{24}}", models.UpdateItemHandler).Methods("POST")
	router.HandleFunc("/upload-image", models.UploadImageHandler).Methods("POST")

	router.HandleFunc("/", models.ProductsHandler)
	router.HandleFunc("/headers", others.HeadersHandler)
	router.HandleFunc("/list", models.ListHandler)

	// обработчик для отображения страницы регистрации (GET)
	router.HandleFunc("/registration", auth.ShowRegistrationPage)
	router.HandleFunc("/register", auth.RegisterHandler) // обработчик для страницы регистрации
	// страница входа и её обработчики
	router.HandleFunc("/login", auth.LoginPageHandler).Methods("GET")
	router.HandleFunc("/login", auth.LoginHandler).Methods("POST")
	router.HandleFunc("/logout", auth.LogoutHandler)      // Exit
	router.HandleFunc("/administrator", admin.AdminPanel) // admin panel
	router.HandleFunc("/administrator/{id}", admin.DeleteUser).Methods("DELETE")
	router.HandleFunc("/tt", models.Tttt)

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// =======================================================
	authRoutes := router.PathPrefix("/").Subrouter()
	authRoutes.Use(AuthMiddleware)

	// группа маршрутов, требующих авторизации
	authRoutes.HandleFunc("/account", models.Account)                // profile
	authRoutes.HandleFunc("/add-to-cart", models.AddToCartHandler)   // для добавления товара в корзину
	authRoutes.HandleFunc("/users-orders", models.UserOrdersHandler) /* доделать html */
	authRoutes.HandleFunc("/submit_order", models.SubmitOrderHandler).Methods("POST")
	authRoutes.HandleFunc("/update_cart", models.UpdateCartHandler).Methods("POST")
	authRoutes.HandleFunc("/cart", models.ViewCartHandler).Methods("GET")
	authRoutes.HandleFunc("/account/edit", auth.EditProfile)
	authRoutes.HandleFunc("/my-items", models.ListUserSaleItems).Methods("GET")
	authRoutes.HandleFunc("/delete-item/{id}", models.DeleteItem).Methods("DELETE")

	// Настройка обработчика для статических файлов
	// router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	filePath := filepath.Join(".", "web", "static", strings.TrimPrefix(r.URL.Path, "/static/"))
	// 	ext := filepath.Ext(filePath)
	// 	mimeType := mime.TypeByExtension(ext)
	// 	if mimeType != "" {
	// 		w.Header().Set("Content-Type", mimeType)
	// 	}
	// 	http.ServeFile(w, r, filePath)
	// })))

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
