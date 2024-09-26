package main

import (
	"log"

	//_ "test/1.7.2/proxy/docs"
	_ "test/docs"
)

// @title ГеоAPI
// @version 1.0
// @description Поиск информации по адресу или координатам

// @host localhost:8080
// @BasePath /

func main() {
	log.Fatal(NewServer(":8080", "hugo", "1313").Start())
}
