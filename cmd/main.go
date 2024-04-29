// - Как видеть, что соединение с базой "ок" и ошибок нету? db.Ping() не помогает.
// Хочется при запуске сервака, что бы ПРОВЕРЯЛ и писал что все "ок"
// V Авторизация клиентов (и ЛК?)
// - Страницу с корзиной клиента
// - сделай хеширование при авторизации и регистрации
// - Дизайн сайта начать добавлять
// V Рефакторинг соединение с БД
// V Подключение БД через метод сделай
// V Распредели по файлам код архитектурно
// V Сделай переходы между страницами или временную навигацию
// - Проверка аутентификации пользователя: Убедитесь, что пользователь 
// аутентифицирован, прежде чем разрешить добавление товара в корзину. Вы можете 
// использовать сеансы, токены или другие методы аутентификации.
// - Проверка наличия товара: Перед добавлением товара в корзину проверьте его 
// наличие в базе данных и его актуальность (например, наличие на складе).
// - Обработка количества товаров: Учет количества добавляемых товаров в корзину. 
// Возможно, вам нужно будет обновлять количество товаров, если товар уже 
// присутствует в корзине, вместо добавления новой записи.
// V "github.com/lib/pq" из файла "DB.go" для чего нужен?
// - добавление заказов в корзину с последующим сохранением в БД
// - токены для пользователя JWT
// - добавь имя зарегистрировавшегося на сайте где то наверху, что бы было ясно, 
// кого обслуживается на сайте.


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
