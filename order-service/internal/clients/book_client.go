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

type BookClient struct {
	baseURL    string
	httpClient *http.Client
	logger     Logger
}

// Logger interface matching the logger implementation
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
}

func NewBookClient(baseURL string, timeout time.Duration, logger Logger) *BookClient {
	return &BookClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

func (c *BookClient) GetBook(ctx context.Context, id int) (*models.BookInfo, error) {
	url := fmt.Sprintf("%s/books/%d", c.baseURL, id)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to get book", "id", id, "error", err)
		return nil, fmt.Errorf("book service unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("Book service returned error", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("book not found")
	}

	var book models.BookInfo
	if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &book, nil
}

func (c *BookClient) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("book service health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("book service unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
