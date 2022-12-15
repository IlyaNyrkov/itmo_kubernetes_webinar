package main

import (
	"app/router"
	"log"
	"net/http"
)

func main() {
	r := router.Router()
	log.Printf("Starting server on the port %s...\n", "8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
