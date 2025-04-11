package config

import (
	"os"

	"github.com/joho/godotenv"
)

func GetPort() string {

	err := godotenv.Load()

	if err != nil {
		return ""
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // Default port if not specified
	}
	return port
}
