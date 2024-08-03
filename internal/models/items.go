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

// Specification представляет спецификации продукта
// type Specification struct {
// 	Brand       string `json:"brand"`
// 	Model       string `json:"model"`
// 	ScreenSize  string `json:"screenSize"`
// 	Resolution  string `json:"resolution"`
// 	RefreshRate string `json:"refreshRate"`
// 	SmartTV     bool   `json:"smartTV"`
// }

// Review представляет отзыв на продукт
type Review struct {
	User    string  `json:"user"`
	Rating  float64 `json:"rating"`
	Comment string  `json:"comment"`
}

// Item представляет структуру товара
type Item struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	User_ID       int                `json:"user_id"`
	Name          string             `json:"name"`
	Price         float64            `json:"price"`
	DynamicFields []DynamicField     `json:"dynamic_fields"`
}

type DynamicField struct {
	FieldName  string `json:"field_name"`
	FieldValue string `json:"field_value"`
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
func getItemFields(itemID string) ([]DynamicField, error) {
	item, err := getItemByID(itemID)
	if err != nil {
		return nil, err
	}

	return item.DynamicFields, nil
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

	fields, err := getItemFields(id)
	if err != nil {
		http.Error(w, "Unable to fetch fields", http.StatusInternalServerError)
		return
	}

	data := struct {
		UserName string
		Item     Item
        Fields   []DynamicField
	}{
		UserName: userName,
		Item:     item,
		Fields:   fields,
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
		// Category: "",
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

	collection := database.Client.Database("myDatabase").Collection("products")
	_, err = collection.UpdateByID(context.Background(), item.ID, bson.M{"$set": item})
	if err != nil {
		http.Error(w, "Error updating database", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/item/%s", item.ID.Hex()), http.StatusSeeOther)
}
