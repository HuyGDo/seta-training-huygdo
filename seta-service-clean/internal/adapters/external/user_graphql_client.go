package external

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"seta/internal/application/ports"
	"seta/internal/domain/common"
	"time"

	"github.com/google/uuid"
)

// GQLAuthService is an adapter that implements the AuthService port.
// It communicates with an external GraphQL user service.
type GQLAuthService struct {
	userServiceURL string
	client         *http.Client
}

// NewGQLAuthService creates a new instance of the GraphQL authentication service client.
func NewGQLAuthService(url string) ports.AuthService {
	return &GQLAuthService{
		userServiceURL: url,
		client:         &http.Client{Timeout: time.Second * 10},
	}
}

// doGQLRequest is a helper function to execute a GraphQL request and decode the response.
func (s *GQLAuthService) doGQLRequest(ctx context.Context, query map[string]interface{}, target interface{}) error {
	jsonQuery, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("failed to create graphql query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.userServiceURL, bytes.NewBuffer(jsonQuery))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to user service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("user service returned non-200 status: %d", resp.StatusCode)
	}

	// The target interface must be a pointer for json.NewDecoder to work.
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode user service response: %w", err)
	}

	return nil
}

func (s *GQLAuthService) ValidateToken(ctx context.Context, token string) (*ports.User, error) {
	query := map[string]interface{}{
		"query": `
            query VerifyToken($token: String!) {
                verifyToken(token: $token) {
                    success
                    user { userId role }
                }
            }
        `,
		"variables": map[string]interface{}{"token": token},
	}

	var result struct {
		Data struct {
			VerifyToken struct {
				Success bool `json:"success"`
				User    struct {
					UserID string `json:"userId"`
					Role   string `json:"role"`
				} `json:"user"`
			} `json:"verifyToken"`
		} `json:"data"`
		Errors []map[string]interface{} `json:"errors"`
	}

	if err := s.doGQLRequest(ctx, query, &result); err != nil {
		return nil, err
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %v", result.Errors[0]["message"])
	}
	if !result.Data.VerifyToken.Success {
		return nil, errors.New("token is not valid")
	}

	userID, err := uuid.Parse(result.Data.VerifyToken.User.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID from auth service: %w", err)
	}

	return &ports.User{
		ID:   common.UserID(userID),
		Role: result.Data.VerifyToken.User.Role,
	}, nil
}

func (s *GQLAuthService) GetUser(ctx context.Context, userID common.UserID) (*ports.User, error) {
	query := map[string]interface{}{
		"query": `
            query GetUser($userId: ID!) {
                user(userId: $userId) {
                    userId
                    role
                }
            }
        `,
		"variables": map[string]interface{}{"userId": userID.String()},
	}

	var result struct {
		Data struct {
			User *struct {
				UserID string `json:"userId"`
				Role   string `json:"role"`
			} `json:"user"`
		} `json:"data"`
		Errors []map[string]interface{} `json:"errors"`
	}

	if err := s.doGQLRequest(ctx, query, &result); err != nil {
		return nil, err
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %v", result.Errors[0]["message"])
	}
	if result.Data.User == nil {
		return nil, errors.New("user not found")
	}

	uid, err := uuid.Parse(result.Data.User.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID from auth service: %w", err)
	}

	return &ports.User{
		ID:   common.UserID(uid),
		Role: result.Data.User.Role,
	}, nil
}

func (s *GQLAuthService) IsLeadManager(ctx context.Context, teamID common.TeamID, userID common.UserID) (bool, error) {
	// TODO: Implement this method by querying the user service GraphQL API.
	// This will likely involve creating a new GraphQL query to check if a user is a lead manager of a specific team.
	return false, errors.New("IsLeadManager not implemented")
}

func (s *GQLAuthService) IsTeamManager(ctx context.Context, userID common.UserID, teamID common.TeamID) (bool, error) {
	// TODO: Implement this method by querying the user service GraphQL API.
	// This will likely involve creating a new GraphQL query to check if a user is a manager of a specific team.
	return false, errors.New("IsTeamManager not implemented")
}
