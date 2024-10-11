package others

import (
	"context"
	"log"
	"net/http"

	"github.com/KrepkiyOrex/inter_store/inventory"
	"google.golang.org/grpc"
)

func NewInventoryClient() (inventory.InventoryServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial("localhost:9000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
		return nil, nil, err
	}
	client := inventory.NewInventoryServiceClient(conn)
	return client, conn, nil
}

func AddInventory(ctx context.Context, client inventory.InventoryServiceClient, productID, warehouseID string, quantity int32) (*inventory.AddInventoryResponse, error) {
	return client.AddInventory(ctx, &inventory.AddInventoryRequest{
		ProductId:   productID,
		WarehouseId: warehouseID,
		Quantity:    quantity,
	})
}

// =====================================================================================================

// TestInventoryHandler обрабатывает запросы для проверки микросервиса
func TestInventoryHandler(w http.ResponseWriter, r *http.Request) {
	client, conn, err := NewInventoryClient()
	if err != nil {
		http.Error(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
		defer conn.Close()

	/* 
		сперва для оформления сделай, потом для страницы с товаром.
		это именно добавляет или убавляет все таки кол товара на складе?
		Переменные уже сможешь подставить сразу сюда, за место product123 и warehouse456 
		и прочее, для декремента. сама функция будет срабатывать, только после нажатия
		 кнопки "оформить заказ", засунешь в ту функ.

		с GetInventory как получить тут инфу о наличии на складе? у меня пока так
		 работает, что отправляет инфу на сервак и там все завершается, но нету 
		 обратного ответа для вывода на фронт для карты.

		*кстати по поводу соедниения, как будет выглядеть код для интернет соединеня,
		 на случай если будет не на хосте у меня а в инете где нить на хостинг?
	*/

	// Пример вызова метода AddInventory
	resp, err := AddInventory(r.Context(), client, "product123", "warehouse456", 10)
	if err != nil {
		http.Error(w, "Failed to call AddInventory", http.StatusInternalServerError)
		return
	}

	if !resp.GetSuccess() {
		http.Error(w, "Failed to add inventory", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Inventory added successfully"))
}
