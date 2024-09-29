package main

import (
	//_ "test/1.7.2/proxy/docs"
	_ "test/docs"

	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

// @title ГеоAPI
// @version 1.0
// @description Поиск информации по адресу или координатам.

// @host localhost:8080
// @BasePath /

func main() {
	NewServer(":8080", "hugo", "1313").Serve()
}
