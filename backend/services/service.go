package services

import (
	"backend/models"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// func GetUrlValues(code string) url.Values {
// 	data := url.Values{}
// 	data.Set("grant_type", "authorization_code")
// 	data.Set("client_id", os.Getenv("PROCORE_CLIENT_ID"))
// 	data.Set("client_secret", os.Getenv("PROCORE_CLIENT_SECRET"))
// 	data.Set("code", code)
// 	data.Set("redirect_uri", "urn:ietf:wg:oauth:2.0:oob")
// 	return data
// }

// DecodeRequestBody decodes the incoming request body into the given struct
func DecodeRequestBody(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// BuildTokenRequestData prepares URL-encoded form data for token exchange
func BuildTokenRequestData(code string) url.Values {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", os.Getenv("PROCORE_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("PROCORE_CLIENT_SECRET"))
	data.Set("code", code)
	data.Set("redirect_uri", "urn:ietf:wg:oauth:2.0:oob")
	return data
}

// FetchAuthToken calls the Procore token endpoint and parses the response
func FetchAuthToken(code string) (*models.AuthTokenResponse, error) {
	reqURL := os.Getenv("TOCKEN_URL")
	data := BuildTokenRequestData(code)

	client := &http.Client{Timeout: 10 * time.Second}
	request, err := http.NewRequest("POST", reqURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, errors.New("failed to create token request")
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)
	if err != nil {
		return nil, errors.New("failed to call token endpoint")
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New("failed to read token response")
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	var tokenResp models.AuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, errors.New("failed to parse token response")
	}

	return &tokenResp, nil
}
