package models

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/KrepkiyOrex/inter_store/internal/auth"
	"github.com/KrepkiyOrex/inter_store/internal/database"
	"github.com/KrepkiyOrex/inter_store/internal/utils"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// отзыв на продукт
// type Review struct {
// 	User    string  `json:"user"`
// 	Rating  float64 `json:"rating"`
// 	Comment string  `json:"comment"`
// }

// Item представляет структуру товара
type Item struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	User_ID           int                `bson:"user_id" json:"user_id"`
	Name              string             `bson:"name" json:"name"`
	Price             float64            `bson:"price" json:"price"`
	Quantity          int32              `bson:"quantity" json:"quantity"`
	ImageURL          string             `bson:"imageURL,omitempty"`
	DynamicFields     []DynamicField     `bson:"dynamic_fields" json:"dynamic_fields"`
	DescriptionFields []DescriptionField `bson:"description_fields" json:"description_fields"`
}

type DynamicField struct {
	FieldName  string `json:"field_name"`
	FieldValue string `json:"field_value"`
}

type DescriptionField struct {
	NameDep  string `bson:"field_name"`
	ValueDep string `bson:"field_value"`
}

// getItemByID получает документ товара из MongoDB по ObjectID
func getItemByID(id string) (Item, error) {
	// Получаем базу данных и коллекцию
	collection := database.GetCollection()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return Item{}, err
	}

	var item Item
	err = collection.FindOne(context.Background(), bson.M{"_id": oid}).Decode(&item)
	if err != nil {
		return Item{}, err
	}
	return item, nil
}

// getItemFields получает все динамические поля товара по его ID
func getItemFields(itemID string) ([]DynamicField, []DescriptionField, error) {
	item, err := getItemByID(itemID)
	if err != nil {
		return nil, nil, err
	}

	return item.DynamicFields, item.DescriptionFields, nil
}

// shows the card of a specific product
func HandlerItemRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	userName, _ := auth.GetUserName(r)

	item, err := getItemByID(id)
	if err != nil {
		http.Error(w, "Unable to fetch data", http.StatusInternalServerError)
		return
	}

	fieldsDin, fieldsDep, err := getItemFields(id)
	if err != nil {
		http.Error(w, "Unable to fetch fields", http.StatusInternalServerError)
		return
	}

	data := struct {
		UserName  string
		Item      Item
		FieldsDin []DynamicField
		FieldsDep []DescriptionField
	}{
		UserName:  userName,
		Item:      item,
		FieldsDin: fieldsDin,
		FieldsDep: fieldsDep,
	}

	utils.RenderTemplate(w, data,
		"web/html/item.html",
		"web/html/navigation.html")
}

// =========================================================================

// create new item
func CreateNewItemHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromCookie(r)
	if err != nil {
		http.Error(w, "User ID not found: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Создание нового пустого товара
	newItem := Item{
		ID:       primitive.NewObjectID(),
		User_ID:  userID,
		Name:     "Edit name",
		Price:    0.0,
		Quantity: 0,
	}

	log.Printf("Item: %v", newItem)

	// сохранение товара в MongoDB
	collection := database.GetCollection()
	_, err = collection.InsertOne(context.Background(), newItem)
	if err != nil {
		http.Error(w, "Error saving to database", http.StatusInternalServerError)
		return
	}

	// перенаправление на страницу редактирования нового товара
	http.Redirect(w, r, fmt.Sprintf("/edit-item/%s", newItem.ID.Hex()), http.StatusSeeOther)
}

// edit item
func EditItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Получаем объект товара из базы данных
	item, err := getItemByID(id)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	log.Printf("Item data: %+v", item)

	data := struct {
		Item Item
	}{
		Item: item,
	}

	// Отправляем данные на страницу редактирования
	utils.RenderTemplate(w, data,
		"web/html/edit_item.html",
		"web/html/navigation.html")
}

// update data item
func UpdateItemHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	item, err := getItemByID(id)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	// обновление данных
	item.Name = r.FormValue("name")
	item.Price, _ = strconv.ParseFloat(r.FormValue("price"), 64)
	// item.Quantity, _ = strconv.ParseInt(r.FormValue("quantity"), 10, 32)
	quantity, _ := strconv.ParseInt(r.FormValue("quantity"), 10, 32)
	item.Quantity = int32(quantity)

	// Получение пути к изображению
	imageURL := r.FormValue("imageURL")
	if imageURL != "" {
		item.ImageURL = imageURL
		log.Printf("ImageURL: %v", item.ImageURL)
	}

	// обработка динамических полей
	dynamicFields := []DynamicField{}
	fieldNames := r.Form["field-name"]
	fieldValues := r.Form["field-value"]
	log.Printf("fieldNames: %v", fieldNames)   // Лог для отладки
	log.Printf("fieldValues: %v", fieldValues) // Лог для отладки
	for i := 0; i < len(fieldNames); i++ {
		dynamicFields = append(dynamicFields, DynamicField{
			FieldName:  fieldNames[i],
			FieldValue: fieldValues[i],
		})
	}
	item.DynamicFields = dynamicFields

	// обработка динамических полей описания
	descriptionFields := []DescriptionField{}
	nameDeps := r.Form["field-name-dep"]
	valueDeps := r.Form["field-value-dep"]
	log.Printf("NameDep: %v", nameDeps)   // Лог для отладки
	log.Printf("ValueDep: %v", valueDeps) // Лог для отладки
	for i := 0; i < len(nameDeps); i++ {
		descriptionFields = append(descriptionFields, DescriptionField{
			NameDep:  nameDeps[i],
			ValueDep: valueDeps[i],
		})
	}
	item.DescriptionFields = descriptionFields

	log.Printf("Dinamic: %v", item.DynamicFields)
	log.Printf("Descrip: %v", item.DescriptionFields)
	log.Printf("ImageURL: %v", item.ImageURL)

	collection := database.GetCollection()
	_, err = collection.UpdateByID(context.Background(), item.ID, bson.M{"$set": item})
	if err != nil {
		http.Error(w, "Error updating database", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/item/%s", item.ID.Hex()), http.StatusSeeOther)
}

// shows the sale items created by the specified user
func ListUserSaleItems(w http.ResponseWriter, r *http.Request) {
	userName, err := auth.GetUserName(r)
	if err != nil {
		renderLoginPage(w)
		return
	}

	userID, err := auth.GetCookieUserID(w, r)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}
	fmt.Println("Parsed User ID:", userID)

	items, err := getUserSaleItems(userID)
	if err != nil {
		http.Error(w, "Failed to find items", http.StatusInternalServerError)
		return
	}

	renderUserSaleItemsPage(w, userName, items)
}

// deleting item from DB list
func DeleteItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := getItemIDFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	objID, err := convertToObjectID(itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = deleteItemFromDatabase(objID)
	if err != nil {
		http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		return
	}

	sendSuccessResponse(w)
}

// ========================================================

func UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("itemImage")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// делаем уникальное имя
	uniqueFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(header.Filename))
	filePath := filepath.Join("web", "static", "img", uniqueFileName)

	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	// путь к файлу для сохранения
	imageURL := "/static/img/" + uniqueFileName

	// отправляем путь к файлу обратно клиенту
	w.Write([]byte(imageURL))
}

// ==========================================================================================

// ========================================================

func renderLoginPage(w http.ResponseWriter) {
	utils.RenderTemplate(w, UserCookie{},
		"web/html/login.html",
		"web/html/navigation.html")
}

func getUserSaleItems(userID int) ([]Item, error) {
	collection := database.GetCollection()

	// создаем фильтр для поиска документов по user_id
	filter := bson.M{"user_id": userID}

	// Выполняем запрос к коллекции
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// Читаем документы из курсора
	var items []Item
	if err := cursor.All(context.Background(), &items); err != nil {
		return nil, err
	}

	return items, nil
}

func renderUserSaleItemsPage(w http.ResponseWriter, userName string, items []Item) {
	data := struct {
		UserName string
		Items    []Item
	}{
		UserName: userName,
		Items:    items,
	}

	utils.RenderTemplate(w, data,
		"web/html/my_items.html",
		"web/html/navigation.html")
}

func getItemIDFromRequest(r *http.Request) (string, error) {
	params := mux.Vars(r)
	itemID := params["id"]
	if itemID == "" {
		return "", fmt.Errorf("Item ID is required")
	}
	return itemID, nil
}

func convertToObjectID(itemID string) (primitive.ObjectID, error) {
	objID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("Invalid item ID")
	}
	return objID, nil
}

func deleteItemFromDatabase(objID primitive.ObjectID) error {
	collection := database.GetCollection()
	filter := bson.M{"_id": objID}
	_, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}
	log.Printf("Item deleted successfully: %v", objID.Hex())
	return nil
}

func sendSuccessResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Item deleted successfully")
}

// Структура для хранения данных о погоде
// type WeatherData struct {
// 	DataCurrent struct {
// 		Temperature float64 `json:"temperature"`
// 		Humidity    int     `json:"humidity"`
// 		Condition   string  `json:"condition"`
// 	} `json:"data_current"`
// }

func Tttt(w http.ResponseWriter, r *http.Request) {
	// Получаем имя пользователя
	userName, err := auth.GetUserName(r)
	if err != nil {
		utils.RenderTemplate(w, nil, "web/html/list.html", "web/html/navigation.html")
		return
	}

	// Запрос к API MeteoBlue
	// resp, err := http.Get("https://my.meteoblue.com/packages/basic-1h_basic-day_current?apikey=OZKqv7rTz8SuNwDI&lat=54.3282&lon=48.3866&asl=176&format=json")
	// if err != nil {
	// 	http.Error(w, "Не удалось получить данные о погоде", http.StatusInternalServerError)
	// 	return
	// }
	// defer resp.Body.Close()

	// // // Чтение и парсинг JSON-ответа
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	http.Error(w, "Не удалось прочитать данные о погоде", http.StatusInternalServerError)
	// 	return
	// }

	// // fmt.Println(string(body))

	// var weatherData WeatherData
	// err = json.Unmarshal(body, &weatherData)
	// if err != nil {
	// 	http.Error(w, "Не удалось разобрать данные о погоде", http.StatusInternalServerError)
	// 	return
	// }

	// // Проверка данных о погоде
	var weatherInfo string
	// if weatherData.DataCurrent.Temperature == 0 && weatherData.DataCurrent.Humidity == 0 && weatherData.DataCurrent.Condition == "" {
	// 	weatherInfo = "Данные о погоде временно недоступны."
	// } else {
	// 	weatherInfo = fmt.Sprintf("Температура: %.1f°C, Влажность: %d%%, Условия: %s",
	// 		weatherData.DataCurrent.Temperature,
	// 		weatherData.DataCurrent.Humidity,
	// 		weatherData.DataCurrent.Condition)
	// }

	// Подготовка данных для шаблона
	data := struct {
		UserName string
		Weather  string
	}{
		UserName: userName,
		Weather:  weatherInfo,
	}

	// Рендеринг шаблона
	utils.RenderTemplate(w, data,
		"web/html/my.html",
		"web/html/navigation.html")
}
