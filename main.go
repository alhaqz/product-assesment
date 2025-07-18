package main

import (
	"be-assesment-product/config"
	"be-assesment-product/db"
	"be-assesment-product/handler"
	"be-assesment-product/redis"
	"be-assesment-product/router"
	"be-assesment-product/service"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

var appConfig *config.Config

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	appConfig = config.InitConfig()

	gormDB, err := config.InitDB(appConfig.Dsn)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	dbProvider := db.NewProvider(gormDB)

	rdb := redis.NewRedis(appConfig.RedisAddr, appConfig.RedisPass, appConfig.RedisDB)

	logger := config.NewLogger()

	productService := service.NewProductService(dbProvider, logger, rdb)

	productHandler := handler.NewProductHandler(productService)

	r := router.NewRouter(productHandler)

	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Server error:", err)
	}
}
