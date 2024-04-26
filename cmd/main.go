// - Как видеть, что соединение с базой "ок" и ошибок нету? db.Ping() не помогает.
// Хочется при запуске сервака, что бы ПРОВЕРЯЛ и писал что все "ок"
// V Авторизация клиентов (и ЛК?)
// - сделай шифрование при авторизации и регистрации
// - Страницу с корзиной клиента
// - Дизайн сайта начать добавлять
// - Рефакторинг соединение с БД
// - Подключение БД через метод сделай
// V Распредели по файлам код архитектурно
// - Сделай переходы между страницами или временную навигацию
// "github.com/lib/pq" из файла "DB.go" для чего нужен?

// {{ template "navigation.html" . }}

package main

import (
	"log"

	"github.com/gorilla/sessions"

	"First_internet_store/internal/api"
	// "First_internet_store/internal/db"
	"First_internet_store/internal/utils"
)

func main() {
	// db.Connect()

	// Инициализируем хранилище сессий
	utils.Store = sessions.NewCookieStore([]byte("your-secret-key"))

	// log.Println("Connecting to the database successfully")
	log.Println("Server starts")

	api.StartServer()

	/*
			Это более короткий способ?

			// Запускаем сервер
		    log.Fatal(http.ListenAndServe(":8080", nil))
	*/
}

// func userDashboardHandler(w http.ResponseWriter, r *http.Request) {

// 	log.Println("Handling user dashboard request")

// 	session, err := store.Get(r, "session-name")
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Проверяем, авторизован ли пользователь
// 	if session.Values["user_id"] == nil {
// 		// Если пользователь не авторизован, перенаправляем на страницу входа
// 		http.Redirect(w, r, "/login", http.StatusSeeOther)
// 		return
// 	}

// 	// Получаем имя пользователя из сессии
// 	username := session.Values["user_id"].(string)

// 	// Выводим приветствие
// 	fmt.Fprintf(w, "Welcome, %s!", username)
// }
