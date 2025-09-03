package external

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"seta/internal/application/ports"
	"strconv"
	"sync"
	"time"
)

// userJob now includes a line number for better error tracking.
type userJob struct {
	lineNumber int
	record     []string
}

// jobResult now contains enough detail to report specific errors.
type jobResult struct {
	success    bool
	lineNumber int
	record     []string
	message    string
}

// GQLUserImporter is an adapter that implements the UserImporter port.
// It orchestrates the process of importing users by calling an external GraphQL service.
type GQLUserImporter struct {
	userServiceURL string
}

// NewGQLUserImporter creates a new instance of the GraphQL user importer.
func NewGQLUserImporter(url string) *GQLUserImporter {
	return &GQLUserImporter{userServiceURL: url}
}

// ImportUsers orchestrates the entire CSV import process.
func (importer *GQLUserImporter) ImportUsers(ctx context.Context, file io.Reader) (ports.UserImportSummary, error) {
	reader := csv.NewReader(file)

	// Read header
	if _, err := reader.Read(); err != nil {
		if err == io.EOF {
			return ports.UserImportSummary{}, nil
		}
		return ports.UserImportSummary{}, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Determine number of workers
	numWorkers := 10
	if v, err := strconv.Atoi(os.Getenv("USER_IMPORT_WORKERS")); err == nil && v > 0 {
		numWorkers = v
	}

	jobs := make(chan userJob)
	results := make(chan jobResult, numWorkers*2) // Buffered channel
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go importer.worker(ctx, jobs, results, &wg)
	}

	// Goroutine to close the results channel once all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Feed jobs from the CSV file
	go func() {
		defer close(jobs)
		line := 1 // Start after header
		for {
			line++
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			// We handle malformed rows in the result collection phase
			if err != nil {
				results <- jobResult{success: false, lineNumber: line, record: []string{"malformed row"}, message: err.Error()}
				continue
			}

			select {
			case <-ctx.Done():
				return // Stop feeding if context is cancelled
			case jobs <- userJob{lineNumber: line, record: record}:
			}
		}
	}()

	// Collect results
	summary := ports.UserImportSummary{Failures: make([]ports.FailedRecord, 0)}
	for r := range results {
		if r.success {
			summary.Succeeded++
		} else {
			summary.Failed++
			summary.Failures = append(summary.Failures, ports.FailedRecord{
				Record: r.record,
				Reason: fmt.Sprintf("Line %d: %s", r.lineNumber, r.message),
			})
		}
	}

	return summary, nil
}

// worker processes jobs from the jobs channel.
func (importer *GQLUserImporter) worker(ctx context.Context, jobs <-chan userJob, results chan<- jobResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		select {
		case <-ctx.Done():
			results <- jobResult{success: false, lineNumber: job.lineNumber, record: job.record, message: "Request canceled"}
			continue
		default:
			err := importer.callCreateUserMutation(ctx, job.record)
			if err != nil {
				results <- jobResult{success: false, lineNumber: job.lineNumber, record: job.record, message: err.Error()}
			} else {
				results <- jobResult{success: true, lineNumber: job.lineNumber, record: job.record, message: "User created"}
			}
		}
	}
}

// callCreateUserMutation sends a GraphQL mutation with retries and context handling.
func (importer *GQLUserImporter) callCreateUserMutation(ctx context.Context, record []string) error {
	if len(record) < 4 {
		return fmt.Errorf("invalid record: not enough columns")
	}

	payload := map[string]any{
		"query": `mutation CreateUser($input: CreateUserInput!) {
                    createUser(input: $input) { success errors }
                  }`,
		"variables": map[string]any{
			"input": map[string]any{
				"username": record[0],
				"email":    record[1],
				"password": record[2],
				"role":     record[3],
			},
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal query: %w", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, importer.userServiceURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			if attempt == maxRetries {
				return fmt.Errorf("user service connection error after %d attempts: %w", maxRetries, err)
			}
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}

		// Process response in a closure to handle defer
		respErr := func() error {
			defer resp.Body.Close()
			if resp.StatusCode >= 400 {
				body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
				return fmt.Errorf("user service HTTP %d: %s", resp.StatusCode, string(body))
			}

			var result struct {
				Data struct {
					CreateUser struct {
						Success bool     `json:"success"`
						Errors  []string `json:"errors"`
					} `json:"createUser"`
				} `json:"data"`
				Errors []struct {
					Message string `json:"message"`
				} `json:"errors"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}
			if len(result.Errors) > 0 {
				return fmt.Errorf("GraphQL error: %s", result.Errors[0].Message)
			}
			if !result.Data.CreateUser.Success {
				return fmt.Errorf("API error: %v", result.Data.CreateUser.Errors)
			}
			return nil
		}()

		if respErr == nil {
			return nil // Success
		}
		if attempt == maxRetries {
			return respErr // Return last error
		}
		time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
	}
	return fmt.Errorf("unexpected error in retry loop")
}
