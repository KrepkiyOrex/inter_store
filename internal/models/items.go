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
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	User_ID int                `json:"user_id"`
	Name    string             `json:"name"`
	Price   float64            `json:"price"`
	// Category         string             `json:"category"`
	// Specifications   Specification      `json:"specifications"`
	// Reviews          []Review        `json:"reviews"`
	// AboutTheProducts AboutTheProduct `json:"aboutTheProducts"`
	Fields map[string]string `json:"fields"` // предполагаемое поле для динамических данных
}

type AboutTheProduct struct {
	Position1 string `json:"position1"`
	Position2 string `json:"position2"`
	Position3 string `json:"position3"`
	Position4 string `json:"position4"`
	Position5 string `json:"position5"`
}

type Field struct {
	Name  string `json:"field_name"`
	Value string `json:"field_value"`
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

func getItemFields(itemID string) ([]Field, error) {
	item, err := getItemByID(itemID)
	if err != nil {
		return nil, err
	}

	fields := []Field{}
	for key, value := range item.Fields {
		fields = append(fields, Field{
			Name:  key,
			Value: value,
		})
	}

	return fields, nil
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
		Fields   []Field
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
	vars := mux.Vars(r)
	id := vars["id"]

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	updateFields := make(map[string]interface{})
	for key, values := range r.Form {
		if len(values) > 0 {
			value := values[0]
			log.Printf("Field name: %s, Field value: %s", key, value)
			if key == "price" {
				price, err := strconv.ParseFloat(value, 64)
				if err != nil {
					http.Error(w, "Invalid price value", http.StatusBadRequest)
					return
				}
				updateFields["price"] = price
			} else {
				updateFields[key] = value
			}
		}
	}

	log.Printf("Update fields: %v", updateFields)

	collection := database.Client.Database("myDatabase").Collection("products")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	update := bson.M{
		"$set": updateFields,
	}
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": oid}, update)
	if err != nil {
		http.Error(w, "Error updating the item", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/item/%s", id), http.StatusSeeOther)
}
