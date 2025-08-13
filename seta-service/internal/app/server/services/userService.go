package services

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
)

// Summary contains the final results of the import process.
type Summary struct {
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
}

// userJob represents a single line from the CSV to be processed.
type userJob struct {
	record []string
}

// jobResult holds the outcome of processing a single userJob.
type jobResult struct {
	success bool
	message string
}

// UserService handles the business logic for user-related operations.
type UserService struct{}

// NewUserService creates a new instance of UserService.
func NewUserService() *UserService {
	return &UserService{}
}

// ImportUsers orchestrates the entire CSV import process.
func (s *UserService) ImportUsers(file io.Reader) (Summary, error) {
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return Summary{}, fmt.Errorf("failed to parse CSV file: %w", err)
	}

	numJobs := len(records) - 1 // Subtract header row
	if numJobs <= 0 {
		return Summary{}, nil // Return empty summary if no data rows
	}

	jobs := make(chan userJob, numJobs)
	results := make(chan jobResult, numJobs)
	var wg sync.WaitGroup

	numWorkers := 10 // This can be made configurable
	for w := 0; w < numWorkers; w++ {
		go s.worker(jobs, results, &wg)
	}

	for _, record := range records[1:] { // Skip header
		wg.Add(1)
		jobs <- userJob{record: record}
	}
	close(jobs)

	wg.Wait()
	close(results)

	summary := Summary{}
	for result := range results {
		if result.success {
			summary.Succeeded++
		} else {
			summary.Failed++
		}
	}
	return summary, nil
}

// worker processes jobs from the jobs channel.
func (s *UserService) worker(jobs <-chan userJob, results chan<- jobResult, wg *sync.WaitGroup) {
	for job := range jobs {
		err := s.callCreateUserMutation(job.record)
		if err != nil {
			results <- jobResult{success: false, message: err.Error()}
		} else {
			results <- jobResult{success: true, message: "User created"}
		}
	}
}

// callCreateUserMutation sends a GraphQL mutation to the user-service.
func (s *UserService) callCreateUserMutation(record []string) error {
	userServiceURL := os.Getenv("USER_SERVICE_URL")
    
	if userServiceURL == "" {
        userServiceURL = "http://localhost:4000/users" // Default for local dev
    }
	
	if len(record) < 4 {
		return fmt.Errorf("invalid record: not enough columns")
	}

	userInput := gin.H{
		"username": record[0],
		"email":    record[1],
		"password": record[2],
		"role":     record[3],
	}

	query := gin.H{
		"query": `
            mutation CreateUser($input: CreateUserInput!) {
              createUser(input: $input) { success, errors }
            }`,
		"variables": gin.H{"input": userInput},
	}

	jsonData, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("failed to marshal query: %w", err)
	}

	resp, err := http.Post(userServiceURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("user service connection error: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			CreateUser struct {
				Success bool     `json:"success"`
				Errors  []string `json:"errors"`
			} `json:"createUser"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Data.CreateUser.Success {
		return fmt.Errorf("API error: %v", result.Data.CreateUser.Errors)
	}

	return nil
}