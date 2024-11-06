package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/KrepkiyOrex/inter_store/internal/api"
	"github.com/KrepkiyOrex/inter_store/internal/utils"
)

func main() {
	// Инициализируем хранилище сессий
	// utils.Store = sessions.NewCookieStore([]byte("your-secret-key"))

	utils.InitLogrus(log.DebugLevel, false, true)

	log.Info("Server starts")

	api.StartServer()
}
