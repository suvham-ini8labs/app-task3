package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"order-service/internal/models"
)

type UserClient struct {
	baseURL    string
	httpClient *http.Client
	logger     Logger
}

func NewUserClient(baseURL string, timeout time.Duration, logger Logger) *UserClient {
	return &UserClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

func (c *UserClient) GetUser(ctx context.Context, id int, token string) (*models.UserInfo, error) {
	url := fmt.Sprintf("%s/users/%d", c.baseURL, id)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add authorization header if token is provided
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to get user", "id", id, "error", err)
		return nil, fmt.Errorf("user service unavailable: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Error("Failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("User service returned error", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("user not found")
	}

	var user models.UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}

func (c *UserClient) ValidateToken(ctx context.Context, token string) (*models.UserInfo, error) {
	// For now, we'll use the token to get user info by making a request to user service
	// In a real implementation, you'd have a dedicated validate endpoint
	// Since user service doesn't have a validate endpoint, we'll use the token to get user info
	// This is a simplified approach
	
	// For demo purposes, we'll extract the user ID from the token
	// In production, you'd properly validate with the user service
	c.logger.Debug("Validating token", "token", token[:min(10, len(token))]+"...")
	
	// Mock implementation - in production, you'd call user service's validate endpoint
	// For now, we'll return a mock user
	return &models.UserInfo{
		ID:       1,
		Username: "john",
		Email:    "john@example.com",
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (c *UserClient) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("user service health check failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Error("Failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("user service unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
