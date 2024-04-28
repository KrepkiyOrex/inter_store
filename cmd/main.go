// - Как видеть, что соединение с базой "ок" и ошибок нету? db.Ping() не помогает.
// Хочется при запуске сервака, что бы ПРОВЕРЯЛ и писал что все "ок"
// V Авторизация клиентов (и ЛК?)
// - Страницу с корзиной клиента
// - сделай шифрование при авторизации и регистрации
// - Дизайн сайта начать добавлять
// V Рефакторинг соединение с БД
// V Подключение БД через метод сделай
// V Распредели по файлам код архитектурно
// V Сделай переходы между страницами или временную навигацию
// "github.com/lib/pq" из файла "DB.go" для чего нужен?

// {{ template "navigation.html" . }}

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
