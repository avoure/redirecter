package db

import (
	"fmt"
	"log"
	"os"
	"redirecter/pkg/models"

	"github.com/joho/godotenv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init() *gorm.DB {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error while getting .env file. Err: %s", err)
	}

	var POSTGRES_USER string = os.Getenv("POSTGRES_USER")
	var POSTGRES_PASSWORD string = os.Getenv("POSTGRES_PASSWORD")
	var POSTGRES_DB string = os.Getenv("POSTGRES_DB")
	var POSTGRES_PORT string = os.Getenv("POSTGRES_PORT")

	dbURL := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s", POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_PORT, POSTGRES_DB)
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}

	db.AutoMigrate(&models.RedirectMap{})

	return db
}
