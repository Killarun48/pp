package main

// @title API
// @version 1.0
// @description API для взаимодействия с пользователями.

// @host localhost:8080
// @BasePath /

func main() {
	NewServer(":8080").Serve()
}
