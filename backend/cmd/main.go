package main

import (
	"log"
	"os"
	"time"

	"backend/internal/database"
	"backend/internal/models"
	"backend/internal/routes"

	"github.com/joho/godotenv"
)

func main() {
	// Try loading .env variables
	_ = godotenv.Load()

	database.Connect()

	database.DB.AutoMigrate(
		&models.User{},s
		&models.Event{},
		&models.Registration{},
	)

	// Background job to delete expired events
	go func() {
		for {
			database.DB.Where("event_date <= ?", time.Now()).Delete(&models.Event{})
			time.Sleep(1 * time.Minute)
		}
	}()

	r := routes.SetupRouter()
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Server failed: ", err)
	}
}