package handlers

import (
	"backend/models"
	"backend/services"
	"encoding/json"
	"fmt"
	"net/http"
)

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World")
}
func GetAuthToken(w http.ResponseWriter, r *http.Request) {
	var req models.AuthTokenRequest
	if err := services.DecodeRequestBody(r, &req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		http.Error(w, `{"error":"Authorization code is required"}`, http.StatusBadRequest)
		return
	}

	tokenResp, err := services.FetchAuthToken(req.Code)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": tokenResp.AccessToken,
		"token_type":   tokenResp.TokenType,
		"expires_in":   tokenResp.ExpiresIn,
	})
}
