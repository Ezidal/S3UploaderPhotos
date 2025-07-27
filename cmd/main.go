package main

import (
	handlers "Ezidal/YandexOSUploaderPhotos/internal/handlers"
	"Ezidal/YandexOSUploaderPhotos/internal/storage"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	log.Printf("Loading environment variables from .env file")
	_ = godotenv.Load()
	_ = storage.LoadStorage()

	log.Printf("Starting Gin server...")

	gin.SetMode(gin.ReleaseMode)
	log.Printf("Setting Gin mode to Release")

	log.Printf("Creating a new Gin router")
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/img", "./img")

	log.Printf("Setting up routes")
	r.GET("/ping", handlers.PingHandler)
	r.GET("/", handlers.MainPage)
	r.POST("/upload", handlers.Upload)

	log.Printf("Starting server on port 8080")
	r.Run()
}
