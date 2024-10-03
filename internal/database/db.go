package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"

	"github.com/pelletier/go-toml" // Пакет для работы с файлами TOML
)

type DBConfig struct {
	Driver string `toml:"driver"`
	DSN    string `toml:"dsn"`
}

type Config struct {
	Database DBConfig `toml:"database"`
}

// DB представляет структуру базы данных
type DB struct {
	*sql.DB
}

// postgreSQL DB
func Connect() (*DB, error) {
	// Load the configuration from the file
	cfg, err := toml.LoadFile("config/config.toml")
	if err != nil {
		log.Println("Error loading configuration:", err)
	}

	// Getting parameters of connection to the database from the configuration
	dbConfig := cfg.Get("database").(*toml.Tree)
	driver := dbConfig.Get("driver").(string)
	dsn := dbConfig.Get("dsn").(string)

	// Connect to the database
	db, err := sql.Open(driver, dsn)
	if err != nil {
		log.Println("Error connecting to the database")
		return nil, err
	}
	// defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Println("Error pinging the database")
		return nil, err
	}

	log.Print("Connecting to the database successfully: ")

	return &DB{DB: db}, nil
}
