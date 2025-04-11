package main

import (
	"backend/config"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello World")
	})
	port := config.GetPort()

	log.Fatal(http.ListenAndServe(":"+port, r))
}
