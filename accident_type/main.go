package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Load environment variables from .env file
func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

// AccidentLog represents the structure of a Procore accident log
type AccidentLog struct {
	ID           int         `json:"id"`
	Comments     string      `json:"comments"`
	Date         string      `json:"date"`
	CreatedAt    string      `json:"created_at"`
	CreatedBy    CreatedBy   `json:"created_by"`
	Status       string      `json:"status"`
	Attachments  []string    `json:"attachments"`
	CustomFields interface{} `json:"custom_fields"`
}

// CreatedBy represents the user who created the accident log
type CreatedBy struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
}

// AccidentTypeResponse represents the API response with accident type
type AccidentTypeResponse struct {
	AccidentLogID int    `json:"accident_log_id"`
	AccidentType  string `json:"accident_type"`
	Date          string `json:"date"`
	ReportedBy    string `json:"reported_by"`
}

// Config holds the application configuration
type Config struct {
	ProjectID string
	CompanyID string
	APIToken  string
}

func main() {
	// Load configuration
	config := Config{
		ProjectID: os.Getenv("PROCORE_PROJECT_ID"),
		CompanyID: os.Getenv("PROCORE_COMPANY_ID"),
		APIToken:  os.Getenv("PROCORE_API_TOKEN"),
	}

	// Initialize Gin router
	r := gin.Default()

	// Define API endpoint
	r.GET("/accidents", func(c *gin.Context) {
		// Get start and end dates from query parameters
		startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, -3, 0).Format("2006-01-02"))
		endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

		// Fetch accident logs from Procore
		logs, err := fetchAccidentLogs(config, startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Process logs to extract accident types
		results := make([]AccidentTypeResponse, 0)
		for _, log := range logs {
			accidentType := extractAccidentType(log.Comments)
			if accidentType != "" {
				results = append(results, AccidentTypeResponse{
					AccidentLogID: log.ID,
					AccidentType:  accidentType,
					Date:          log.Date,
					ReportedBy:    log.CreatedBy.Name,
				})
			}
		}

		c.JSON(http.StatusOK, results)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}

// fetchAccidentLogs retrieves accident logs from Procore API
func fetchAccidentLogs(config Config, startDate, endDate string) ([]AccidentLog, error) {
	client := &http.Client{}
	url := fmt.Sprintf("https://sandbox.procore.com/rest/v1.0/projects/%s/accident_logs", config.ProjectID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	req.Header.Add("Procore-Company-Id", config.CompanyID)
	req.Header.Add("Authorization", "Bearer "+config.APIToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Add query parameters
	q := req.URL.Query()
	q.Add("start_date", startDate)
	q.Add("end_date", endDate)
	req.URL.RawQuery = q.Encode()

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("procore API returned status: %d", resp.StatusCode)
	}

	var logs []AccidentLog
	if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
		return nil, err
	}

	return logs, nil
}

// extractAccidentType parses the accident type from the comments field
func extractAccidentType(comments string) string {
	re := regexp.MustCompile(`\[Type: ([^\]]+)\]`)
	matches := re.FindStringSubmatch(comments)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
