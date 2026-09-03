package config

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/mohammed-ayoub-dz/hifz/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

var dbOnce sync.Once

func Db() {
	dbOnce.Do(func() {

		_ = godotenv.Load()

		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		sslmode := os.Getenv("DB_SSLMODE")

		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			host,
			user,
			password,
			dbname,
			port,
			sslmode,
		)

		db, err := gorm.Open(
			postgres.Open(dsn),
			&gorm.Config{},
		)

		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatal("Failed to get SQL database:", err)
		}

		sqlDB.SetMaxOpenConns(5)
		sqlDB.SetMaxIdleConns(2)

		DB = db

		err = DB.AutoMigrate(
			&models.User{},
			&models.HifzProgress{},
			&models.DailyProgress{},
			&models.HifzSession{},
		)

		if err != nil {
			log.Fatal("Failed to migrate database:", err)
		}

		fmt.Println("Successfully connected to the database.")
	})
}