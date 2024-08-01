package models

import (
	"First_internet_store/internal/auth"
	"First_internet_store/internal/database"
	"First_internet_store/internal/utils"
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Specification представляет спецификации продукта
type Specification struct {
	Brand       string `json:"brand"`
	Model       string `json:"model"`
	ScreenSize  string `json:"screenSize"`
	Resolution  string `json:"resolution"`
	RefreshRate string `json:"refreshRate"`
	SmartTV     bool   `json:"smartTV"`
}

// Review представляет отзыв на продукт
type Review struct {
	User    string  `json:"user"`
	Rating  float64 `json:"rating"`
	Comment string  `json:"comment"`
}

// Item представляет структуру товара
type Item struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name             string             `json:"name"`
	Category         string             `json:"category"`
	Price            float64            `json:"price"`
	Specifications   Specification      `json:"specifications"`
	Reviews          []Review           `json:"reviews"`
	AboutTheProducts AboutTheProduct    `json:"aboutTheProducts"`
}

type AboutTheProduct struct {
	Position1 string `json:"position1"`
	Position2 string `json:"position2"`
	Position3 string `json:"position3"`
	Position4 string `json:"position4"`
	Position5 string `json:"position5"`
}

// getItemByID получает документ товара из MongoDB по ObjectID
func getItemByID(id string) (Item, error) {
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

	data := struct {
		UserName string
		Item     Item
	}{
		UserName: userName,
		Item:     item,
	}

	utils.RenderTemplate(w, data,
		"web/html/item.html",
		"web/html/navigation.html")
}

// =========================================================================

// =========================================================================

// create new item
func CreateNewItemHandler(w http.ResponseWriter, r *http.Request) {
	// Создание нового пустого товара
	newItem := Item{
		ID: primitive.NewObjectID(),
		// Начальные значения можно задать пустыми строками или дефолтными значениями
		Name:     "",
		Category: "",
		Price:    0.0,
	}

	// Сохранение товара в MongoDB
	collection := database.Client.Database("myDatabase").Collection("products")
	_, err := collection.InsertOne(context.Background(), newItem)
	if err != nil {
		http.Error(w, "Error saving to database", http.StatusInternalServerError)
		return
	}

	// Перенаправление на страницу редактирования нового товара
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

	// Отправляем данные на страницу редактирования
	utils.RenderTemplate(w, item,
		"web/html/edit_item.html",
		"web/html/navigation.html")
}

func UpdateItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Парсинг данных из формы
	name := r.FormValue("name")
	category := r.FormValue("category")
	priceStr := r.FormValue("price")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		http.Error(w, "Invalid price format", http.StatusBadRequest)
		return
	}

	// Обновление товара в базе данных
	collection := database.Client.Database("myDatabase").Collection("products")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	update := bson.M{
		"$set": bson.M{
			"name":     name,
			"Category": category,
			"price":    price,
		},
	}
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": oid}, update)
	if err != nil {
		http.Error(w, "Error updating the item", http.StatusInternalServerError)
		return
	}

	// Перенаправление на страницу просмотра товара после обновления
	http.Redirect(w, r, fmt.Sprintf("/item/%s", id), http.StatusSeeOther)
}
