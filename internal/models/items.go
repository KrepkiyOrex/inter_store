package models

import (
	"First_internet_store/internal/auth"
	"First_internet_store/internal/database"
	"First_internet_store/internal/utils"
	"context"
	"fmt"
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
	DynamicFields     []DynamicField     `bson:"dynamic_fields" json:"dynamic_fields"`
	DescriptionFields []DescriptionField `bson:"description_fields" json:"description_fields"`
}

type DynamicField struct {
	FieldName  string `json:"field_name"`
	FieldValue string `json:"field_value"`
}

type DescriptionField struct {
	NameDep  string `json:"field_name"`
	ValueDep string `json:"field_value"`
}

// getItemByID получает документ товара из MongoDB по ObjectID
func getItemByID(id string) (Item, error) {
	// Получаем базу данных и коллекцию
	collection := database.Client.Database("myDatabase").Collection("products")
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

// обрабатывает запросы к MongoDB
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

	utils.RenderTemplate(w, data, "web/html/item.html", "web/html/navigation.html")
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
	collection := database.Client.Database("myDatabase").Collection("products")
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

	collection := database.Client.Database("myDatabase").Collection("products")
	_, err = collection.UpdateByID(context.Background(), item.ID, bson.M{"$set": item})
	if err != nil {
		http.Error(w, "Error updating database", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/item/%s", item.ID.Hex()), http.StatusSeeOther)
}

func ListUserSaleItems(w http.ResponseWriter, r *http.Request) {
	userName, err := auth.GetUserName(r)
	if err != nil {
		// Куки не найдено, показываем форму входа
		utils.RenderTemplate(w, UserCookie{},
			"web/html/login.html",
			"web/html/navigation.html")
		return
	}

	// Получение идентификатора пользователя из куки
	cookieID, err := r.Cookie("userID")
	if err != nil || cookieID.Value == "" {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(cookieID.Value)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}
	fmt.Println("Parsed User ID:", userID) // Отладочное сообщение

	collection := database.Client.Database("myDatabase").Collection("products")

	// Создаем фильтр для поиска документов по user_id
	filter := bson.M{"user_id": userID}

	// Выполняем запрос к коллекции
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		http.Error(w, "Failed to find items", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	// Читаем документы из курсора
	var items []Item
	if err := cursor.All(context.Background(), &items); err != nil {
		http.Error(w, "Failed to read items", http.StatusInternalServerError)
		return
	}

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

// deleting item from DB list
func DeleteItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	itemID := params["id"]
	if itemID == "" {
		http.Error(w, "Item ID is required", http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	collection := database.Client.Database("myDatabase").Collection("products")
	filter := bson.M{"_id": objID}
	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Item deleted successfully")
}
