package main

import (
	"backend/config"
	"backend/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/home", handlers.HelloWorld)
	r.HandleFunc("/auth", handlers.GetAuthToken)
	port := config.GetPort()

	log.Fatal(http.ListenAndServe(":"+port, r))
}
