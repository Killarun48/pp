package main

import (
	"log"
)

func main() {
	log.Fatal(NewServer(":8080", "hugo", "1313").Start())
}
