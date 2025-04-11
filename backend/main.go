package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello World")
	})
	err := godotenv.Load()
	if err != nil {
		return
	}
	port := os.Getenv("PORT")
	fmt.Println("port:", port)
	if port == "" {
		port = "8000" // Default port if not specified
	}

	fmt.Printf("Server running at http://localhost:%s/home\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
