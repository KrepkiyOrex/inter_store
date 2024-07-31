package models

import (
	"First_internet_store/internal/auth"
	"First_internet_store/internal/database"
	"First_internet_store/internal/utils"
	"context"
	"net/http"

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
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string             `json:"name"`
	Category       string             `json:"category"`
	Price          float64            `json:"price"`
	Specifications Specification      `json:"specifications"`
	Reviews        []Review           `json:"reviews"`
}

// getItemByID получает документ из MongoDB по ObjectID
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

// MongoHandler обрабатывает запросы к MongoDB
func MongoHandler(w http.ResponseWriter, r *http.Request) {
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

	utils.RenderTemplate(w, data, "web/html/list2.html", "web/html/navigation.html")
}
