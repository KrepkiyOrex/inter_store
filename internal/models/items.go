package models

import (
	"First_internet_store/internal/auth"
	"First_internet_store/internal/database"
	"First_internet_store/internal/utils"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Review представляет отзыв на продукт
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
	Image             string             `bson:"image,omitempty"`
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
		ID:      primitive.NewObjectID(),
		User_ID: userID,
		Name:    "Edit name",
		Price:   0.0,
	}

	log.Printf("Item: %v", newItem)

	// Сохранение товара в MongoDB
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

	// Обновление данных
	item.Name = r.FormValue("name")
	item.Price, _ = strconv.ParseFloat(r.FormValue("price"), 64)

	// Обработка динамических полей
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

	// Обработка динамических полей описания
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
	// Парсинг данных формы
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Чтение данных файла
	tempFile, err := ioutil.TempFile("uploads", "upload-*.png")
	if err != nil {
		http.Error(w, "Error creating temp file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	tempFile.Write(fileBytes)

	// Получение itemID из формы
	itemID := r.FormValue("itemID")

	// Обновление пути к изображению в базе данных
	err = updateItemImage(itemID, tempFile.Name())
	if err != nil {
		http.Error(w, "Failed to update item image", http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully uploaded file: %s\n", handler.Filename)
}

func updateItemImage(itemID string, imagePath string) error {
	objID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		return err
	}

	collection := database.GetCollection()
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"image": imagePath}}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	return err
}

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
