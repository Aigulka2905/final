package server

import (
	"final/pkg/api"
	"log"
	"net/http"
	"os"
)

func Run() error {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}
	api.Init() // Register API handlers
	http.Handle("/", http.FileServer(http.Dir("./web")))
	log.Printf("Server starting on port %s", port)
	return http.ListenAndServe(":"+port, nil)
}
