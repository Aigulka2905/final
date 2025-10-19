package main

import (
	"final/pkg/db"
	"final/pkg/server"
	"log"
)

func main() {
	if err := db.Init("scheduler.db"); err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
