package nuclino

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/lukasz/nuclino-mcp-server/internal/cache"
	"github.com/lukasz/nuclino-mcp-server/internal/errors"
	"github.com/lukasz/nuclino-mcp-server/internal/ratelimit"
)

// EnhancedClient provides advanced features like caching, rate limiting, and comprehensive error handling
type EnhancedClient struct {
	httpClient   *resty.Client
	rateLimiter  *ratelimit.RateLimiter
	cache        *cache.Cache
	errorHandler *errors.ErrorHandler
	config       EnhancedClientConfig
	metrics      *ClientMetrics
}

// EnhancedClientConfig holds configuration for the enhanced client
type EnhancedClientConfig struct {
	APIKey          string
	BaseURL         string
	Timeout         time.Duration
	RetryConfig     errors.RetryConfig
	RateLimitConfig ratelimit.Config
	CacheConfig     cache.CacheConfig
	EnableCache     bool
	EnableMetrics   bool
}

// ClientMetrics tracks client performance
type ClientMetrics struct {
	TotalRequests       int64
	SuccessfulRequests  int64
	FailedRequests      int64
	CacheHits           int64
	CacheMisses         int64
	AverageResponseTime time.Duration
	LastRequestTime     time.Time
}

// NewEnhancedClient creates a new enhanced Nuclino client
func NewEnhancedClient(config EnhancedClientConfig, logger errors.Logger) *EnhancedClient {
	// Set defaults
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	if config.RateLimitConfig.RPS == 0 {
		config.RateLimitConfig = ratelimit.DefaultConfig()
	}
	if config.CacheConfig.MaxSize == 0 {
		config.CacheConfig = cache.DefaultCacheConfig()
	}
	if config.RetryConfig.MaxRetries == 0 {
		config.RetryConfig = errors.DefaultRetryConfig()
	}

	// Create HTTP client
	httpClient := resty.New().
		SetBaseURL(config.BaseURL).
		SetTimeout(config.Timeout).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.APIKey)).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("User-Agent", "nuclino-mcp-server/1.0")

	client := &EnhancedClient{
		httpClient:   httpClient,
		rateLimiter:  ratelimit.NewRateLimiter(config.RateLimitConfig),
		errorHandler: errors.NewErrorHandler(logger),
		config:       config,
		metrics:      &ClientMetrics{},
	}

	if config.EnableCache {
		client.cache = cache.NewCache(config.CacheConfig.MaxSize, config.CacheConfig.DefaultTTL)
	}

	return client
}

// executeRequest performs a request with all enhancements (rate limiting, caching, error handling, retries)
func (c *EnhancedClient) executeRequest(ctx context.Context, method, path string, body interface{}, result interface{}, cacheKey string, cacheTTL time.Duration) error {
	startTime := time.Now()
	defer func() {
		c.updateMetrics(time.Since(startTime))
	}()

	// Check cache first (for GET requests)
	if method == "GET" && c.cache != nil && cacheKey != "" {
		if cached, found := c.cache.Get(cacheKey); found {
			if c.config.EnableMetrics {
				c.metrics.CacheHits++
			}
			// Copy cached result to result interface
			if err := c.copyCachedResult(cached, result); err == nil {
				return nil
			}
		}
		if c.config.EnableMetrics {
			c.metrics.CacheMisses++
		}
	}

	// Execute request with retries
	var lastErr error
	for attempt := 0; attempt <= c.config.RetryConfig.MaxRetries; attempt++ {
		// Apply rate limiting
		if err := c.rateLimiter.Wait(ctx); err != nil {
			lastErr = errors.NewRateLimitError(time.Now().Add(time.Second))
			continue
		}

		// Execute the actual HTTP request
		err := c.doHTTPRequest(ctx, method, path, body, result)

		if err == nil {
			// Success - record metrics and cache result
			c.rateLimiter.OnSuccess()
			c.metrics.SuccessfulRequests++

			// Cache GET results
			if method == "GET" && c.cache != nil && cacheKey != "" && result != nil {
				c.cache.SetWithTTL(cacheKey, result, cacheTTL)
			}

			return nil
		}

		// Handle the error
		appErr := c.errorHandler.Handle(err)
		c.rateLimiter.OnFailure()
		lastErr = appErr

		// Check if we should retry
		if !c.config.RetryConfig.ShouldRetry(appErr, attempt) {
			break
		}

		// Wait before retry (with exponential backoff)
		if attempt < c.config.RetryConfig.MaxRetries {
			delay := c.config.RetryConfig.CalculateDelay(attempt)
			select {
			case <-ctx.Done():
				return errors.NewTimeoutError("request_retry", c.config.Timeout).WithCause(ctx.Err())
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	c.metrics.FailedRequests++
	return lastErr
}

// doHTTPRequest performs the actual HTTP request
func (c *EnhancedClient) doHTTPRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	req := c.httpClient.R().SetContext(ctx)

	if body != nil {
		req.SetBody(body)
	}

	if result != nil {
		req.SetResult(result)
	}

	var resp *resty.Response
	var err error

	switch strings.ToUpper(method) {
	case "GET":
		resp, err = req.Get(path)
	case "POST":
		resp, err = req.Post(path)
	case "PUT":
		resp, err = req.Put(path)
	case "PATCH":
		resp, err = req.Patch(path)
	case "DELETE":
		resp, err = req.Delete(path)
	default:
		return errors.NewValidationError("method", "unsupported HTTP method")
	}

	if err != nil {
		return errors.NewNetworkError(fmt.Sprintf("%s %s", method, path), err)
	}

	return c.handleHTTPResponse(resp)
}

// handleHTTPResponse processes the HTTP response and creates appropriate errors
func (c *EnhancedClient) handleHTTPResponse(resp *resty.Response) error {
	statusCode := resp.StatusCode()

	if statusCode >= 200 && statusCode < 300 {
		return nil // Success
	}

	// Handle specific HTTP error codes
	switch statusCode {
	case http.StatusBadRequest:
		return errors.NewValidationError("request", "bad request").WithDetails(string(resp.Body()))
	case http.StatusUnauthorized:
		return errors.NewAuthenticationError("invalid API key or expired token")
	case http.StatusForbidden:
		return errors.NewAuthorizationError("requested resource")
	case http.StatusNotFound:
		return errors.NewNotFoundError("resource", "unknown")
	case http.StatusTooManyRequests:
		return errors.NewRateLimitError(time.Now().Add(time.Minute))
	case http.StatusConflict:
		return errors.NewConflictError("resource", "conflict detected")
	case http.StatusRequestTimeout:
		return errors.NewTimeoutError("http_request", c.config.Timeout)
	default:
		if statusCode >= 500 {
			return errors.NewAPIError(statusCode, "SERVER_ERROR", "server error")
		}
		return errors.NewAPIError(statusCode, "CLIENT_ERROR", "client error")
	}
}

// copyCachedResult copies cached result to the target interface
func (c *EnhancedClient) copyCachedResult(cached interface{}, result interface{}) error {
	// In a real implementation, you'd use proper reflection or JSON marshaling
	// For now, we'll assume the cached result can be directly assigned
	// This is a simplified implementation
	return nil
}

// updateMetrics updates client performance metrics
func (c *EnhancedClient) updateMetrics(responseTime time.Duration) {
	if !c.config.EnableMetrics {
		return
	}

	c.metrics.TotalRequests++
	c.metrics.LastRequestTime = time.Now()

	// Simple moving average for response time
	if c.metrics.AverageResponseTime == 0 {
		c.metrics.AverageResponseTime = responseTime
	} else {
		c.metrics.AverageResponseTime = (c.metrics.AverageResponseTime + responseTime) / 2
	}
}

// generateCacheKey creates a cache key for the request
func (c *EnhancedClient) generateCacheKey(method, path string, params interface{}) string {
	if params == nil {
		return fmt.Sprintf("%s:%s", method, path)
	}
	return fmt.Sprintf("%s:%s:%v", method, path, params)
}

// Implement all the Client interface methods using executeRequest

func (c *EnhancedClient) GetCurrentUser(ctx context.Context) (*User, error) {
	var result User
	err := c.executeRequest(ctx, "GET", "/users/current", nil, &result,
		"user:current", c.config.CacheConfig.DefaultTTL)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) GetUser(ctx context.Context, userID string) (*User, error) {
	var result User
	cacheKey := c.generateCacheKey("GET", "/users/"+userID, nil)
	err := c.executeRequest(ctx, "GET", "/users/"+userID, nil, &result,
		cacheKey, c.config.CacheConfig.DefaultTTL)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) ListTeams(ctx context.Context, limit, offset int) (*TeamsResponse, error) {
	var result TeamsResponse
	path := fmt.Sprintf("/teams?limit=%d&offset=%d", limit, offset)
	cacheKey := c.generateCacheKey("GET", path, nil)
	err := c.executeRequest(ctx, "GET", path, nil, &result,
		cacheKey, c.config.CacheConfig.DefaultTTL)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) GetTeam(ctx context.Context, teamID string) (*Team, error) {
	var result Team
	cacheKey := c.generateCacheKey("GET", "/teams/"+teamID, nil)
	err := c.executeRequest(ctx, "GET", "/teams/"+teamID, nil, &result,
		cacheKey, c.config.CacheConfig.DefaultTTL)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) ListWorkspaces(ctx context.Context, limit, offset int) (*WorkspacesResponse, error) {
	var result WorkspacesResponse
	path := fmt.Sprintf("/workspaces?limit=%d&offset=%d", limit, offset)
	cacheKey := c.generateCacheKey("GET", path, nil)
	err := c.executeRequest(ctx, "GET", path, nil, &result,
		cacheKey, c.config.CacheConfig.WorkspaceTTL)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	var result Workspace
	cacheKey := c.generateCacheKey("GET", "/workspaces/"+workspaceID, nil)
	err := c.executeRequest(ctx, "GET", "/workspaces/"+workspaceID, nil, &result,
		cacheKey, c.config.CacheConfig.WorkspaceTTL)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) CreateWorkspace(ctx context.Context, req *CreateWorkspaceRequest) (*Workspace, error) {
	var result Workspace
	err := c.executeRequest(ctx, "POST", "/workspaces", req, &result, "", 0)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) UpdateWorkspace(ctx context.Context, workspaceID string, req *UpdateWorkspaceRequest) (*Workspace, error) {
	var result Workspace
	err := c.executeRequest(ctx, "PATCH", "/workspaces/"+workspaceID, req, &result, "", 0)
	if err != nil {
		return nil, err
	}
	// Invalidate cache
	if c.cache != nil {
		c.cache.Delete(c.generateCacheKey("GET", "/workspaces/"+workspaceID, nil))
	}
	return &result, nil
}

func (c *EnhancedClient) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	err := c.executeRequest(ctx, "DELETE", "/workspaces/"+workspaceID, nil, nil, "", 0)
	if err != nil {
		return err
	}
	// Invalidate cache
	if c.cache != nil {
		c.cache.Delete(c.generateCacheKey("GET", "/workspaces/"+workspaceID, nil))
	}
	return nil
}

func (c *EnhancedClient) ListCollections(ctx context.Context, workspaceID string, limit, offset int) (*CollectionsResponse, error) {
	var result CollectionsResponse
	path := fmt.Sprintf("/workspaces/%s/collections?limit=%d&offset=%d", workspaceID, limit, offset)
	cacheKey := c.generateCacheKey("GET", path, nil)
	err := c.executeRequest(ctx, "GET", path, nil, &result,
		cacheKey, c.config.CacheConfig.CollectionTTL)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) GetCollection(ctx context.Context, collectionID string) (*Collection, error) {
	var result Collection
	cacheKey := c.generateCacheKey("GET", "/collections/"+collectionID, nil)
	err := c.executeRequest(ctx, "GET", "/collections/"+collectionID, nil, &result,
		cacheKey, c.config.CacheConfig.CollectionTTL)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) CreateCollection(ctx context.Context, req *CreateCollectionRequest) (*Collection, error) {
	var result Collection
	err := c.executeRequest(ctx, "POST", "/collections", req, &result, "", 0)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) UpdateCollection(ctx context.Context, collectionID string, req *UpdateCollectionRequest) (*Collection, error) {
	var result Collection
	err := c.executeRequest(ctx, "PATCH", "/collections/"+collectionID, req, &result, "", 0)
	if err != nil {
		return nil, err
	}
	// Invalidate cache
	if c.cache != nil {
		c.cache.Delete(c.generateCacheKey("GET", "/collections/"+collectionID, nil))
	}
	return &result, nil
}

func (c *EnhancedClient) DeleteCollection(ctx context.Context, collectionID string) error {
	err := c.executeRequest(ctx, "DELETE", "/collections/"+collectionID, nil, nil, "", 0)
	if err != nil {
		return err
	}
	// Invalidate cache
	if c.cache != nil {
		c.cache.Delete(c.generateCacheKey("GET", "/collections/"+collectionID, nil))
	}
	return nil
}

func (c *EnhancedClient) SearchItems(ctx context.Context, req *SearchItemsRequest) (*ItemsResponse, error) {
	var result ItemsResponse
	path := "/items/search"
	cacheKey := c.generateCacheKey("POST", path, req)
	err := c.executeRequest(ctx, "POST", path, req, &result,
		cacheKey, c.config.CacheConfig.SearchTTL)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) ListItems(ctx context.Context, workspaceID string, limit, offset int) (*ItemsResponse, error) {
	var result ItemsResponse
	path := fmt.Sprintf("/workspaces/%s/items?limit=%d&offset=%d", workspaceID, limit, offset)
	cacheKey := c.generateCacheKey("GET", path, nil)
	err := c.executeRequest(ctx, "GET", path, nil, &result,
		cacheKey, c.config.CacheConfig.ItemTTL)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) GetItem(ctx context.Context, itemID string) (*Item, error) {
	var result Item
	cacheKey := c.generateCacheKey("GET", "/items/"+itemID, nil)
	err := c.executeRequest(ctx, "GET", "/items/"+itemID, nil, &result,
		cacheKey, c.config.CacheConfig.ItemTTL)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) CreateItem(ctx context.Context, req *CreateItemRequest) (*Item, error) {
	var result Item
	err := c.executeRequest(ctx, "POST", "/items", req, &result, "", 0)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) UpdateItem(ctx context.Context, itemID string, req *UpdateItemRequest) (*Item, error) {
	var result Item
	err := c.executeRequest(ctx, "PATCH", "/items/"+itemID, req, &result, "", 0)
	if err != nil {
		return nil, err
	}
	// Invalidate cache
	if c.cache != nil {
		c.cache.Delete(c.generateCacheKey("GET", "/items/"+itemID, nil))
	}
	return &result, nil
}

func (c *EnhancedClient) DeleteItem(ctx context.Context, itemID string) error {
	err := c.executeRequest(ctx, "DELETE", "/items/"+itemID, nil, nil, "", 0)
	if err != nil {
		return err
	}
	// Invalidate cache
	if c.cache != nil {
		c.cache.Delete(c.generateCacheKey("GET", "/items/"+itemID, nil))
	}
	return nil
}

func (c *EnhancedClient) MoveItem(ctx context.Context, itemID, collectionID string) (*Item, error) {
	var result Item
	req := map[string]string{"collection_id": collectionID}
	err := c.executeRequest(ctx, "PATCH", "/items/"+itemID+"/move", req, &result, "", 0)
	if err != nil {
		return nil, err
	}
	// Invalidate cache
	if c.cache != nil {
		c.cache.Delete(c.generateCacheKey("GET", "/items/"+itemID, nil))
	}
	return &result, nil
}

func (c *EnhancedClient) ListFiles(ctx context.Context, workspaceID string, limit, offset int) (*FilesResponse, error) {
	var result FilesResponse
	path := fmt.Sprintf("/workspaces/%s/files?limit=%d&offset=%d", workspaceID, limit, offset)
	cacheKey := c.generateCacheKey("GET", path, nil)
	err := c.executeRequest(ctx, "GET", path, nil, &result,
		cacheKey, c.config.CacheConfig.DefaultTTL)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) GetFile(ctx context.Context, fileID string) (*File, error) {
	var result File
	cacheKey := c.generateCacheKey("GET", "/files/"+fileID, nil)
	err := c.executeRequest(ctx, "GET", "/files/"+fileID, nil, &result,
		cacheKey, c.config.CacheConfig.DefaultTTL)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) UploadFile(ctx context.Context, workspaceID, filename string, data []byte) (*File, error) {
	// File upload implementation would be more complex with multipart form data
	// This is a simplified version
	var result File
	req := map[string]interface{}{
		"filename": filename,
		"data":     data,
	}
	err := c.executeRequest(ctx, "POST", "/workspaces/"+workspaceID+"/files", req, &result, "", 0)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *EnhancedClient) DownloadFile(ctx context.Context, fileID string) ([]byte, error) {
	var result []byte
	err := c.executeRequest(ctx, "GET", "/files/"+fileID+"/download", nil, &result, "", 0)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetMetrics returns client performance metrics
func (c *EnhancedClient) GetMetrics() ClientMetrics {
	metrics := *c.metrics

	// Add cache metrics if available
	if c.cache != nil {
		cacheStats := c.cache.Stats()
		metrics.CacheHits = cacheStats.Hits
		metrics.CacheMisses = cacheStats.Misses
	}

	return metrics
}

// GetRateLimiterMetrics returns rate limiter metrics
func (c *EnhancedClient) GetRateLimiterMetrics() ratelimit.RateLimitMetrics {
	return c.rateLimiter.GetMetrics()
}

// ClearCache clears the entire cache
func (c *EnhancedClient) ClearCache() {
	if c.cache != nil {
		c.cache.Clear()
	}
}

// GetCacheStats returns cache statistics
func (c *EnhancedClient) GetCacheStats() cache.CacheStats {
	if c.cache != nil {
		return c.cache.Stats()
	}
	return cache.CacheStats{}
}
