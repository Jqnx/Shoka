package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

type service struct {
	// db *gorm.DB
	db *pgx.Conn
}

var (
	database   = os.Getenv("BLUEPRINT_DB_DATABASE")
	password   = os.Getenv("BLUEPRINT_DB_PASSWORD")
	username   = os.Getenv("BLUEPRINT_DB_USERNAME")
	port       = os.Getenv("BLUEPRINT_DB_PORT")
	host       = os.Getenv("BLUEPRINT_DB_HOST")
	schema     = os.Getenv("BLUEPRINT_DB_SCHEMA")
	dbInstance *service
)

//func New() *gorm.DB {
//	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable TimeZone=Europe/Brussels", username, password, host, port, database)
//	newLogger := logger.New(
//		log.New(os.Stdout, "\r\n", log.LstdFlags),
//		logger.Config{
//			Colorful: true,
//		})
//	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
//		Logger: newLogger,
//		// FullSaveAssociations: false,
//		TranslateError: true,
//	})
//	if err != nil {
//		log.Fatal("Error connecting to database:", err)
//	}
//	return db
//}

func New() *pgx.Conn {
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable TimeZone=Europe/Brussels", username, password, host, port, database)
	db, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	return db
}
