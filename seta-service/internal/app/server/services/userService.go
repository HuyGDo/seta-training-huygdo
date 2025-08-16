package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// FailedRecord holds information about a CSV record that failed to import.
type FailedRecord struct {
	Record []string `json:"record"`
	Reason string   `json:"reason"`
}

// Summary now includes detailed failure information.
type Summary struct {
	Succeeded int            `json:"succeeded"`
	Failed    int            `json:"failed"`
	Failures  []FailedRecord `json:"failures"`
}

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

// UserService handles the business logic for user-related operations.
type UserService struct{}

// NewUserService creates a new instance of UserService.
func NewUserService() *UserService {
	return &UserService{}
}

// ImportUsers orchestrates the entire CSV import process.
func (s *UserService) ImportUsers(ctx context.Context, file io.Reader) (Summary, error) {
    reader := csv.NewReader(file)

    // Read header
    if _, err := reader.Read(); err != nil {
        if err == io.EOF {
            return Summary{}, nil
        }
        return Summary{}, fmt.Errorf("failed to read CSV header: %w", err)
    }

    // Workers
    numWorkers := 10
    if v, _ := strconv.Atoi(os.Getenv("USER_IMPORT_WORKERS")); v > 0 {
        numWorkers = v
    }

    jobs := make(chan userJob)
    results := make(chan jobResult, numWorkers*2) // buffered so workers don't block
    var wg sync.WaitGroup
    wg.Add(numWorkers)
    for i := 0; i < numWorkers; i++ {
        go s.worker(ctx, jobs, results, &wg)
    }

    // Close results when ALL workers are done
    go func() {
        wg.Wait()
        close(results)
    }()

    summary := Summary{Failures: make([]FailedRecord, 0)}
    // Feed jobs in THIS goroutine (no results writes here)
    line := 1 // header
    for {
        line++
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            // Malformed CSV row: record failure locally (don't send to results)
            summary.Failed++
            summary.Failures = append(summary.Failures, FailedRecord{
                Record: []string{"malformed row"},
                Reason: fmt.Sprintf("Line %d: %v", line, err),
            })
            continue
        }

        select {
        case <-ctx.Done():
            // Stop feeding; let workers drain/exit
            close(jobs)
            // Drain whatever results are pending before returning
            for r := range results {
                if r.success {
                    summary.Succeeded++
                } else {
                    summary.Failed++
                    summary.Failures = append(summary.Failures, FailedRecord{
                        Record: r.record,
                        Reason: fmt.Sprintf("Line %d: %s", r.lineNumber, r.message),
                    })
                }
            }
            return summary, ctx.Err()

        case jobs <- userJob{lineNumber: line, record: record}:
        }
    }
    close(jobs)

    // Collect worker results until results is closed by the waiter goroutine
    for r := range results {
        if r.success {
            summary.Succeeded++
        } else {
            summary.Failed++
            summary.Failures = append(summary.Failures, FailedRecord{
                Record: r.record,
                Reason: fmt.Sprintf("Line %d: %s", r.lineNumber, r.message),
            })
        }
    }

    return summary, nil
}


// worker processes jobs from the jobs channel.
func (s *UserService) worker(ctx context.Context, jobs <-chan userJob, results chan<- jobResult, wg *sync.WaitGroup) {
	defer wg.Done() 
	for job := range jobs {
		if ctx.Err() != nil {
			results <- jobResult{success: false, lineNumber: job.lineNumber, record: job.record, message: "Request canceled"}
			continue
		}
		err := s.callCreateUserMutation(ctx, job.record)
		if err != nil {
			results <- jobResult{success: false, lineNumber: job.lineNumber, record: job.record, message: err.Error()}
		} else {
			results <- jobResult{success: true, lineNumber: job.lineNumber, record: job.record, message: "User created"}
		}
	}
}

// callCreateUserMutation sends a GraphQL mutation with retries and context handling.
func (s *UserService) callCreateUserMutation(ctx context.Context, record []string) error {
    userServiceURL := os.Getenv("USER_SERVICE_URL")
    if userServiceURL == "" {
        userServiceURL = "http://localhost:4000/users"
    }
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
    if err != nil { return fmt.Errorf("failed to marshal query: %w", err) }

    client := &http.Client{ Timeout: 15 * time.Second } // ⬅ timeout
    maxRetries := 3

    for attempt := 1; attempt <= maxRetries; attempt++ {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        req, err := http.NewRequestWithContext(ctx, http.MethodPost, userServiceURL, bytes.NewBuffer(jsonData))
        if err != nil { return err }
        req.Header.Set("Content-Type", "application/json")

        resp, err := client.Do(req)
        if err != nil {
            if attempt == maxRetries { return fmt.Errorf("user service connection error after %d attempts: %w", maxRetries, err) }
            time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
            continue
        }

        func() {
            defer resp.Body.Close()

            if resp.StatusCode >= 400 {
                body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
                err = fmt.Errorf("user service HTTP %d: %s", resp.StatusCode, string(body))
                return
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
            if de := json.NewDecoder(resp.Body).Decode(&result); de != nil {
                err = fmt.Errorf("failed to decode response: %w", de); return
            }
            if len(result.Errors) > 0 {
                err = fmt.Errorf("GraphQL error: %s", result.Errors[0].Message); return
            }
            if !result.Data.CreateUser.Success {
                err = fmt.Errorf("API error: %v", result.Data.CreateUser.Errors); return
            }
            err = nil
        }()

        if err == nil { return nil }
        if attempt == maxRetries { return err }
        time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
    }
    return fmt.Errorf("unexpected error in retry loop")
}
