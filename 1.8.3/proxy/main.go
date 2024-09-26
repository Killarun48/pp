package main

import (
	"log"

	//_ "test/1.7.2/proxy/docs"
	_ "test/docs"
)

// @title ГеоAPI
// @version 1.0
// @description Поиск информации по адресу или координатам. \nПолученный токен в методе /api/login: поместить в заголовок через меню "Authorize" с префиксом "Bearer ", пример: Bearer token

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apiKey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	log.Fatal(NewServer(":8080", "hugo", "1313").Start())
}
