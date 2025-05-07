package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Load environment variables
func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

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

type CreatedBy struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
}

type AccidentTypeResponse struct {
	AccidentLogID int    `json:"accident_log_id"`
	AccidentType  string `json:"accident_type"`
	Date          string `json:"date"`
	ReportedBy    string `json:"reported_by"`
	Comments      string `json:"comments"`
}

type AccidentSearchRequest struct {
	StartDate    string `form:"start_date"`
	EndDate      string `form:"end_date"`
	AccidentType string `form:"accident_type"`
}

type Config struct {
	ProjectID string
	CompanyID string
	APIToken  string
}

func main() {
	config := Config{
		ProjectID: os.Getenv("PROCORE_PROJECT_ID"),
		CompanyID: os.Getenv("PROCORE_COMPANY_ID"),
		APIToken:  os.Getenv("PROCORE_API_TOKEN"),
	}

	// Verify environment variables
	if config.ProjectID == "" || config.CompanyID == "" || config.APIToken == "" {
		log.Fatal("Missing required environment variables")
	}

	r := gin.Default()

	// Enhanced CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/accidents", func(c *gin.Context) {
		// Parse query parameters
		startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, -3, 0).Format("2006-01-02"))
		endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
		accidentType := c.Query("accident_type")

		// Validate dates
		if _, err := time.Parse("2006-01-02", startDate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
		if _, err := time.Parse("2006-01-02", endDate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}

		// Fetch and process logs
		logs, err := fetchAccidentLogs(config, startDate, endDate)
		if err != nil {
			log.Printf("Procore API error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch accident logs",
				"details": err.Error(),
			})
			return
		}

		// Filter results
		results := filterLogs(logs, accidentType)
		c.JSON(http.StatusOK, results)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on port %s", port)
	log.Fatal(r.Run(":" + port))
}

func filterLogs(logs []AccidentLog, accidentType string) []AccidentTypeResponse {
	results := make([]AccidentTypeResponse, 0)
	for _, log := range logs {
		extractedType := extractAccidentType(log.Comments)
		if extractedType != "" && (accidentType == "" || strings.EqualFold(extractedType, accidentType)) {
			results = append(results, AccidentTypeResponse{
				AccidentLogID: log.ID,
				AccidentType:  extractedType,
				Date:          log.Date,
				ReportedBy:    log.CreatedBy.Name,
				Comments:      log.Comments,
			})
		}
	}
	return results
}

func fetchAccidentLogs(config Config, startDate, endDate string) ([]AccidentLog, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	url := fmt.Sprintf("https://sandbox.procore.com/rest/v1.0/projects/%s/accident_logs", config.ProjectID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Procore-Company-Id", config.CompanyID)
	req.Header.Add("Authorization", "Bearer "+config.APIToken)
	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	q.Add("start_date", startDate)
	q.Add("end_date", endDate)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("procore API returned status: %d", resp.StatusCode)
	}

	var logs []AccidentLog
	if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return logs, nil
}

func extractAccidentType(comments string) string {
	re := regexp.MustCompile(`(?i)\[Type: ([^\]]+)\]`)
	matches := re.FindStringSubmatch(comments)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}
