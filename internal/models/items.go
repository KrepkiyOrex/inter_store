package models

import (
	"context"
	"database/sql"
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

	"github.com/gorilla/sessions"
)

// store is the global variable for the session store
var store = sessions.NewCookieStore([]byte("super-secret-key"))

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
	Quantity          int32              `bson:"quantity" json:"quantity"`
	Name              string             `json:"name"`  // для combineItems
	Price             float64            `json:"price"` // для combineItems
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

// getItemByIDMongo получает документ товара из MongoDB по ObjectID
func getItemByIDMongo(id string) (Item, error) {
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

func getItemsByUserIDPostgre(userID int) ([]ItemPsql, error) {
	db, err := database.Connect() // Подключение к базе данных PostgreSQL
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Подготовка запроса для выборки всех товаров по user_id
	rows, err := db.QueryContext(
		context.Background(),
		"SELECT id, mongo_id, name, price, image_url FROM products WHERE user_id = $1",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	var items []ItemPsql
	for rows.Next() {
		var item ItemPsql
		if err := rows.Scan(&item.ID, &item.Mongo_id, &item.Name, &item.Price, &item.Image_url); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %v", err)
	}

	return items, nil
}

// getItemFields получает все динамические поля товара по его ID
func getItemFields(itemID string) ([]DynamicField, []DescriptionField, error) {
	item, err := getItemByIDMongo(itemID)
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

	item, err := getItemByIDMongo(id)
	if err != nil {
		http.Error(w, "Unable to fetch data", http.StatusInternalServerError)
		return
	}

	fieldsDin, fieldsDep, err := getItemFields(id)
	if err != nil {
		http.Error(w, "Unable to fetch fields", http.StatusInternalServerError)
		return
	}

	// =========================================================================

	itemPostgre, err := getItemByMongoIDPostgre(id)
	if err != nil {
		http.Error(w, "Postgre item not found", http.StatusNotFound)
		return
	}

	// =========================================================================

	data := struct {
		UserName  string
		Item      Item
		FieldsDin []DynamicField
		FieldsDep []DescriptionField
		ItemPsql  ItemPsql
	}{
		UserName:  userName,
		Item:      item,
		FieldsDin: fieldsDin,
		FieldsDep: fieldsDep,
		ItemPsql:  itemPostgre,
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
	newItemMongo := Item{
		ID:      primitive.NewObjectID(),
		User_ID: userID,
		// Name:              "Edit name",
		// Price:             0.0,
		// DynamicFields:     []DynamicField{},
		// DescriptionFields: []DescriptionField{},
		Quantity: 0,
	}

	log.Printf("Item: %v", newItemMongo)

	// сохранение товара в MongoDB
	collection := database.GetCollection()
	_, err = collection.InsertOne(context.Background(), newItemMongo)
	if err != nil {
		http.Error(w, "Error saving to database", http.StatusInternalServerError)
		return
	}

	//======================================================================

	db, err := database.Connect() // postgreSQL

	newItemPostgre := ItemPsql{
		Mongo_id:  newItemMongo.ID.Hex(), // для конвертации ObjectID в string
		UserID:    userID,
		Price:     0,
		Name:      "userID_P_SQL",
		Image_url: "",
	}

	log.Printf("ItemPsql: %v", newItemPostgre)

	err = db.QueryRowContext(
		context.Background(),
		`INSERT INTO products (mongo_id, name, price, image_url)
			 VALUES ($1, $2, $3, $4) RETURNING id`,
		newItemPostgre.Mongo_id, newItemPostgre.Name, newItemPostgre.Price, newItemPostgre.Image_url,
	).Scan(&newItemPostgre.ID)

	if err != nil {
		http.Error(w, "Error saving to PostgreSQL database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//======================================================================

	// перенаправление на страницу редактирования нового товара
	http.Redirect(w, r, fmt.Sprintf("/edit-item/%s", newItemMongo.ID.Hex()), http.StatusSeeOther)
}

// create new item
// func CreateDraftHandler(w http.ResponseWriter, r *http.Request) {
// 	// Получаем сессию
// 	session, err := store.Get(r, "session")
// 	if err != nil {
// 		http.Error(w, "Error getting session: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Получаем UserID из cookies или сессии
// 	userID, err := utils.GetUserIDFromCookie(r)
// 	if err != nil {
// 		http.Error(w, "User ID not found: "+err.Error(), http.StatusUnauthorized)
// 		return
// 	}

// 	// Создаем черновик товара
// 	draftItem := Item{
// 		ID:       primitive.NewObjectID(),
// 		User_ID:  userID,
// 		Name:     "",
// 		Price:    0.0,
// 		Quantity: 0,
// 	}

// 	// Сохраняем черновик в сессии
// 	session.Values["draftItem"] = draftItem
// 	err = session.Save(r, w)
// 	if err != nil {
// 		http.Error(w, "Failed to save session: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Перенаправляем на страницу редактирования черновика
// 	http.Redirect(w, r, "/edit-item/"+draftItem.ID.Hex(), http.StatusSeeOther)
// }

// Редактирование черновика товара
// func EditDraftItemHandler(w http.ResponseWriter, r *http.Request) {
// 	// Получаем сессию
// 	session, err := store.Get(r, "session")
// 	if err != nil {
// 		http.Error(w, "Error getting session: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Получаем черновик товара из сессии
// 	draftItem, ok := session.Values["draftItem"].(Item)
// 	if !ok {
// 		http.Error(w, "Draft item not found", http.StatusNotFound)
// 		return
// 	}

// 	// Отображаем шаблон редактирования с черновиком товара
// 	utils.RenderTemplate(w, draftItem, "web/html/edit_item.html", "web/html/navigation.html")
// }

// edit item
func EditItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// получаем объект товара из базы данных Mongo
	itemMongo, err := getItemByIDMongo(id)
	if err != nil {
		http.Error(w, "Mongo item not found", http.StatusNotFound)
		return
	}
	log.Printf("Mongo item data: %+v", itemMongo)

	// получаем объект товара из базы данных Postgre
	itemPsql, err := getItemByMongoIDPostgre(id)
	if err != nil {
		http.Error(w, "Postgre item not found", http.StatusNotFound)
		return
	}
	log.Printf("Postgre item data: %+v", itemPsql)

	data := struct {
		Item     Item
		ItemPsql ItemPsql
	}{
		Item:     itemMongo,
		ItemPsql: itemPsql,
	}

	// Отправляем данные на страницу редактирования
	utils.RenderTemplate(w, data,
		"web/html/edit_item.html",
		"web/html/navigation.html")
}

// SaveItemHandler - сохраняет редактируемый товар в базу данных
// func SaveItemHandler(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	id := vars["id"]

// 	// Получаем сессию
// 	session, err := store.Get(r, "session")
// 	if err != nil {
// 		http.Error(w, "Error getting session: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Проверяем, есть ли в сессии черновик товара
// 	draftItem, ok := session.Values["draftItem"].(Item)
// 	if !ok || draftItem.ID.Hex() != id {
// 		http.Error(w, "Draft item not found in session", http.StatusNotFound)
// 		return
// 	}

// 	// Получаем данные из формы
// 	name := r.FormValue("name")
// 	priceStr := r.FormValue("price")
// 	quantityStr := r.FormValue("quantity")

// 	// Преобразуем строковые значения в нужные типы
// 	price, err := strconv.ParseFloat(priceStr, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid price value", http.StatusBadRequest)
// 		return
// 	}

// 	quantity, err := strconv.Atoi(quantityStr)
// 	if err != nil {
// 		http.Error(w, "Invalid quantity value", http.StatusBadRequest)
// 		return
// 	}

// 	// Обновляем черновик товара
// 	draftItem.Name = name
// 	draftItem.Price = price
// 	draftItem.Quantity = int32(quantity) // Преобразуем quantity в int32

// 	// Сохраняем черновик в MongoDB
// 	collection := database.GetCollection()
// 	_, err = collection.InsertOne(context.Background(), draftItem)
// 	if err != nil {
// 		http.Error(w, "Error saving to MongoDB: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Сохраняем черновик в PostgreSQL
// 	db, err := database.Connect()
// 	if err != nil {
// 		http.Error(w, "Error connecting to PostgreSQL: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	_, err = db.ExecContext(
// 		context.Background(),
// 		`INSERT INTO products (mongo_id, name, price, quantity)
//         VALUES ($1, $2, $3, $4)`,
// 		draftItem.ID.Hex(), draftItem.Name, draftItem.Price, draftItem.Quantity,
// 	)
// 	if err != nil {
// 		http.Error(w, "Error saving to PostgreSQL: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Удаляем черновик из сессии
// 	delete(session.Values, "draftItem")
// 	session.Save(r, w)

// 	// Перенаправляем на страницу редактирования товара
// 	http.Redirect(w, r, fmt.Sprintf("/edit-item/%s", draftItem.ID.Hex()), http.StatusSeeOther)
// }

func getItemByMongoIDPostgre(mongoID string) (ItemPsql, error) {
	db, err := database.Connect() // Подключение к базе данных PostgreSQL
	if err != nil {
		return ItemPsql{}, err
	}
	defer db.Close()

	var itemPsql ItemPsql
	err = db.QueryRowContext(
		context.Background(),
		"SELECT id, mongo_id, name, price, image_url FROM products WHERE mongo_id = $1",
		mongoID,
	).Scan(&itemPsql.ID, &itemPsql.Mongo_id, &itemPsql.Name, &itemPsql.Price, &itemPsql.Image_url)

	if err != nil {
		if err == sql.ErrNoRows {
			return ItemPsql{}, fmt.Errorf("no item found with mongo_id: %s", mongoID)
		}
		return ItemPsql{}, fmt.Errorf("error fetching item: %v", err)
	}

	return itemPsql, nil
}

type ItemPsql struct {
	ID        int     `db:"id"`
	Mongo_id  string  `db:"mongo_id"`
	UserID    int     `db:"user_id"`
	Name      string  `db:"name"`
	Price     float64 `db:"price"`
	Image_url string  `db:"image_url"`
}

// update data item
func UpdateItemHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(" +++ start item ")

	id := mux.Vars(r)["id"]
	item, err := getItemByIDMongo(id)
	if err != nil {
		http.Error(w, "Mongo item not found", http.StatusNotFound)
		return
	}

	db, err := database.Connect() // postgreSQL
	if err != nil {
		http.Error(w, "Error connecting to PostgreSQL database", http.StatusInternalServerError)
		return
	}

	var itemPsql ItemPsql
	var userID sql.NullInt64 // для обработки NULL в поле user_id

	// проверяем, есть ли уже запись в PostgreSQL для данного товара
	err = db.QueryRowContext(context.Background(),
		`SELECT id, user_id, name, price, image_url FROM products WHERE mongo_id = $1`,
		id).Scan(&itemPsql.ID, &userID, &itemPsql.Name, &itemPsql.Price, &itemPsql.Image_url)

	if err == sql.ErrNoRows {

		fmt.Println(" +++ sql.ErrNoRows", itemPsql, item.Quantity)

		itemPsql.UserID, err = auth.GetCookieUserID(w, r)
		if err != nil {
			http.Error(w, "Invalid user_ID", http.StatusInternalServerError)
			return
		}

		// присваиваем значения полей из формы
		itemPsql.Name = r.FormValue("name")
		itemPsql.Price, _ = strconv.ParseFloat(r.FormValue("price"), 64)
		quantity, _ := strconv.ParseInt(r.FormValue("quantity"), 10, 32)
		item.Quantity = int32(quantity)

		imageURL := r.FormValue("Image_url")
		if imageURL != "" {
			itemPsql.Image_url = imageURL
			log.Printf("Image_url: %v", itemPsql.Image_url)
		}

		// вставка новой записи в PostgreSQL
		err = db.QueryRowContext(
			context.Background(),
			`INSERT INTO products (user_id, name, price, image_url, mongo_id) 
                VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			itemPsql.UserID, itemPsql.Name, itemPsql.Price, itemPsql.Image_url, id,
		).Scan(&itemPsql.ID)

		if err != nil {
			http.Error(w, "Error saving to PostgreSQL database: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

		// Обработка динамических полей
		// dynamicFields := []DynamicField{}
		// fieldNames := r.Form["field-name"]
		// fieldValues := r.Form["field-value"]
		// for i := 0; i < len(fieldNames); i++ {
		// 	dynamicFields = append(dynamicFields, DynamicField{
		// 		FieldName:  fieldNames[i],
		// 		FieldValue: fieldValues[i],
		// 	})
		// }
		// item.DynamicFields = dynamicFields

		// // Обработка динамических полей описания
		// descriptionFields := []DescriptionField{}
		// nameDeps := r.Form["field-name-dep"]
		// valueDeps := r.Form["field-value-dep"]
		// for i := 0; i < len(nameDeps); i++ {
		// 	descriptionFields = append(descriptionFields, DescriptionField{
		// 		NameDep:  nameDeps[i],
		// 		ValueDep: valueDeps[i],
		// 	})
		// }
		// item.DescriptionFields = descriptionFields

		item.DynamicFields, item.DescriptionFields = processDynamicFields(r)

		// collection := database.GetCollection()
		// _, err = collection.UpdateByID(context.Background(), item.ID, bson.M{"$set": item})
		// if err != nil {
		// 	http.Error(w, "Error updating MongoDB database", http.StatusInternalServerError)
		// 	return
		// }

		// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

	} else if err == nil {
		// Если запись найдена — обновляем ее
		itemPsql.UserID, err = auth.GetCookieUserID(w, r)
		if err != nil {
			http.Error(w, "Invalid user_ID", http.StatusInternalServerError)
			return
		}

		// присваиваем значения полей из формы
		itemPsql.Name = r.FormValue("name")
		itemPsql.Price, _ = strconv.ParseFloat(r.FormValue("price"), 64)
		quantity, _ := strconv.ParseInt(r.FormValue("quantity"), 10, 32)
		item.Quantity = int32(quantity)

		imageURL := r.FormValue("Image_url")
		if imageURL != "" {
			itemPsql.Image_url = imageURL
			log.Printf("Image_url: %v", itemPsql.Image_url)
		}

		// если userID был найден как NULL, обрабатываем его корректно
		if userID.Valid {
			itemPsql.UserID = int(userID.Int64)
			// } else {
			// itemPsql.UserID = 0 // или какое-то другое значение по умолчанию
		}

		// обновляем существующую запись в PostgreSQL
		_, err = db.ExecContext(
			context.Background(),
			`UPDATE products SET user_id = $1, name = $2, price = $3, image_url = $4 WHERE mongo_id = $5`,
			itemPsql.UserID, itemPsql.Name, itemPsql.Price, itemPsql.Image_url, id,
		)

		if err != nil {
			http.Error(w, "Error updating PostgreSQL database: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println(" *** else if err")

		// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

		// Обработка динамических полей
		// dynamicFields := []DynamicField{}
		// fieldNames := r.Form["field-name"]
		// fieldValues := r.Form["field-value"]
		// for i := 0; i < len(fieldNames); i++ {
		// 	dynamicFields = append(dynamicFields, DynamicField{
		// 		FieldName:  fieldNames[i],
		// 		FieldValue: fieldValues[i],
		// 	})
		// }
		// item.DynamicFields = dynamicFields

		// // Обработка динамических полей описания
		// descriptionFields := []DescriptionField{}
		// nameDeps := r.Form["field-name-dep"]
		// valueDeps := r.Form["field-value-dep"]
		// for i := 0; i < len(nameDeps); i++ {
		// 	descriptionFields = append(descriptionFields, DescriptionField{
		// 		NameDep:  nameDeps[i],
		// 		ValueDep: valueDeps[i],
		// 	})
		// }
		// item.DescriptionFields = descriptionFields

		item.DynamicFields, item.DescriptionFields = processDynamicFields(r)

		// collection := database.GetCollection()
		// _, err = collection.UpdateByID(context.Background(), item.ID, bson.M{"$set": item})
		// if err != nil {
		// 	http.Error(w, "Error updating MongoDB database", http.StatusInternalServerError)
		// 	return
		// }

		fmt.Println(" [TEST] process func:", item.DynamicFields, item.DescriptionFields)

		// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

	} else {
		http.Error(w, "Error querying PostgreSQL database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// обновление данных в MongoDB
	collection := database.GetCollection()
	_, err = collection.UpdateByID(context.Background(), item.ID, bson.M{"$set": item})
	if err != nil {
		http.Error(w, "Error updating MongoDB database", http.StatusInternalServerError)
		return
	}

	// Редирект после успешного обновления
	http.Redirect(w, r, fmt.Sprintf("/item/%s", item.ID.Hex()), http.StatusSeeOther)
}

// обработка динамических полей
func processDynamicFields(r *http.Request) ([]DynamicField, []DescriptionField) {
	dynamicFields := []DynamicField{}
	fieldNames := r.Form["field-name"]
	fieldValues := r.Form["field-value"]
	for i := 0; i < len(fieldNames); i++ {
		dynamicFields = append(dynamicFields, DynamicField{
			FieldName:  fieldNames[i],
			FieldValue: fieldValues[i],
		})
	}

	descriptionFields := []DescriptionField{}
	nameDeps := r.Form["field-name-dep"]
	valueDeps := r.Form["field-value-dep"]
	for i := 0; i < len(nameDeps); i++ {
		descriptionFields = append(descriptionFields, DescriptionField{
			NameDep:  nameDeps[i],
			ValueDep: valueDeps[i],
		})
	}

	return dynamicFields, descriptionFields
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
	}
	fmt.Println("Parsed User ID:", userID)

	// Получаем товары из Mongo
	mongoItems, err := getUserSaleItemsMongo(userID)
	if err != nil {
		http.Error(w, "Failed to find items from Mongo", http.StatusInternalServerError)
		return
	}
	fmt.Println("Mongo Items:", mongoItems)

	// Получаем товары из PostgreSQL
	postgresItem, err := getItemsByUserIDPostgre(userID)
	if err != nil {
		http.Error(w, "Failed to find item from PostgreSQL", http.StatusInternalServerError)
		return
	}
	fmt.Println("PostgreSQL Items:", postgresItem)

	combinedItems := combineItems(mongoItems, postgresItem)
	fmt.Println("Combined Items:", combinedItems)

	renderUserSaleItemsPage(w, userName, combinedItems)
}

func combineItems(mongoItems []Item, postgresItems []ItemPsql) []Item {
    var combinedItems []Item

    // Создаём мапу для быстрого поиска товаров из Mongo по mongo_id
    mongoItemsMap := make(map[string]Item)
    for _, mongoItem := range mongoItems {
        mongoItemsMap[mongoItem.ID.Hex()] = mongoItem
    }
    
    fmt.Println("Mongo Map:", mongoItemsMap) // Отладочный вывод карты Mongo товаров

    // Соединяем товары PostgreSQL и Mongo
    for _, postgresItem := range postgresItems {
        mongoItem, found := mongoItemsMap[postgresItem.Mongo_id]
        if found {
            // если товар найден в Mongo, объединяем данные
            combinedItems = append(combinedItems, Item{
                ID:               mongoItem.ID,
                User_ID:          postgresItem.UserID,
                Quantity:         mongoItem.Quantity,
                Name:             postgresItem.Name,
                Price:            postgresItem.Price,
                DynamicFields:    mongoItem.DynamicFields,
                DescriptionFields: mongoItem.DescriptionFields,
            })
        } else {
            // если товар не найден в Mongo, пропускаем его
            fmt.Printf("Skipping PostgreSQL item with Mongo_id: %s (not found in Mongo)\n", postgresItem.Mongo_id)
        }
    }

    fmt.Println("Combined Items after merge:", combinedItems) // Вывод объединённых товаров после обработки

    return combinedItems
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
	imageURL := "static/img/" + uniqueFileName

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

func getUserSaleItemsMongo(userID int) ([]Item, error) {
	collection := database.GetCollection()

	filter := bson.M{"user_id": userID}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

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

func Tttt(w http.ResponseWriter, r *http.Request) {
	// Получаем имя пользователя
	userName, err := auth.GetUserName(r)
	if err != nil {
		utils.RenderTemplate(w, nil, "web/html/list.html", "web/html/navigation.html")
		return
	}

	// // Проверка данных о погоде
	var weatherInfo string

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
