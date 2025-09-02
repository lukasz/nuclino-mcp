package nuclino

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"
)

const (
	defaultBaseURL    = "https://api.nuclino.com"
	defaultTimeout    = 30 * time.Second
	defaultRateLimit  = 10 // requests per second
	defaultRateBurst  = 20
	defaultRetryCount = 3
	defaultRetryDelay = 1 * time.Second
)

// Client interface defines the methods for interacting with Nuclino API
type Client interface {
	// User methods
	GetCurrentUser(ctx context.Context) (*User, error)
	GetUser(ctx context.Context, userID string) (*User, error)

	// Team methods
	ListTeams(ctx context.Context, limit, offset int) (*TeamsResponse, error)
	GetTeam(ctx context.Context, teamID string) (*Team, error)

	// Workspace methods
	ListWorkspaces(ctx context.Context, limit, offset int) (*WorkspacesResponse, error)
	GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error)
	CreateWorkspace(ctx context.Context, req *CreateWorkspaceRequest) (*Workspace, error)
	UpdateWorkspace(ctx context.Context, workspaceID string, req *UpdateWorkspaceRequest) (*Workspace, error)
	DeleteWorkspace(ctx context.Context, workspaceID string) error

	// Collection methods
	ListCollections(ctx context.Context, workspaceID string, limit, offset int) (*CollectionsResponse, error)
	GetCollection(ctx context.Context, collectionID string) (*Collection, error)
	CreateCollection(ctx context.Context, req *CreateCollectionRequest) (*Collection, error)
	UpdateCollection(ctx context.Context, collectionID string, req *UpdateCollectionRequest) (*Collection, error)
	DeleteCollection(ctx context.Context, collectionID string) error

	// Item methods
	SearchItems(ctx context.Context, req *SearchItemsRequest) (*ItemsResponse, error)
	ListItems(ctx context.Context, workspaceID string, limit, offset int) (*ItemsResponse, error)
	GetItem(ctx context.Context, itemID string) (*Item, error)
	CreateItem(ctx context.Context, req *CreateItemRequest) (*Item, error)
	UpdateItem(ctx context.Context, itemID string, req *UpdateItemRequest) (*Item, error)
	DeleteItem(ctx context.Context, itemID string) error
	MoveItem(ctx context.Context, itemID, collectionID string) (*Item, error)

	// File methods
	ListFiles(ctx context.Context, workspaceID string, limit, offset int) (*FilesResponse, error)
	GetFile(ctx context.Context, fileID string) (*File, error)
	UploadFile(ctx context.Context, workspaceID, filename string, data []byte) (*File, error)
	DownloadFile(ctx context.Context, fileID string) ([]byte, error)
}

// client implements the Client interface
type client struct {
	httpClient  *resty.Client
	rateLimiter *rate.Limiter
	apiKey      string
	baseURL     string
}

// NewClient creates a new Nuclino API client
func NewClient(apiKey string) Client {
	httpClient := resty.New().
		SetBaseURL(defaultBaseURL).
		SetTimeout(defaultTimeout).
		SetRetryCount(defaultRetryCount).
		SetRetryWaitTime(defaultRetryDelay).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey)).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")

	// Add retry conditions
	httpClient.AddRetryCondition(func(r *resty.Response, err error) bool {
		return r.StatusCode() >= 500 || r.StatusCode() == 429
	})

	rateLimiter := rate.NewLimiter(rate.Limit(defaultRateLimit), defaultRateBurst)

	return &client{
		httpClient:  httpClient,
		rateLimiter: rateLimiter,
		apiKey:      apiKey,
		baseURL:     defaultBaseURL,
	}
}

// NewClientWithConfig creates a new client with custom configuration
func NewClientWithConfig(apiKey, baseURL string, rateLimitRPS int, timeout time.Duration) Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if rateLimitRPS <= 0 {
		rateLimitRPS = defaultRateLimit
	}
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	httpClient := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(timeout).
		SetRetryCount(defaultRetryCount).
		SetRetryWaitTime(defaultRetryDelay).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey)).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")

	httpClient.AddRetryCondition(func(r *resty.Response, err error) bool {
		return r.StatusCode() >= 500 || r.StatusCode() == 429
	})

	rateLimiter := rate.NewLimiter(rate.Limit(rateLimitRPS), rateLimitRPS*2)

	return &client{
		httpClient:  httpClient,
		rateLimiter: rateLimiter,
		apiKey:      apiKey,
		baseURL:     baseURL,
	}
}

func (c *client) makeRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	// Apply rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	req := c.httpClient.R().SetContext(ctx)

	if body != nil {
		req.SetBody(body)
	}

	if result != nil {
		req.SetResult(result)
	}

	var resp *resty.Response
	var err error

	switch method {
	case http.MethodGet:
		resp, err = req.Get(path)
	case http.MethodPost:
		resp, err = req.Post(path)
	case http.MethodPut:
		resp, err = req.Put(path)
	case http.MethodPatch:
		resp, err = req.Patch(path)
	case http.MethodDelete:
		resp, err = req.Delete(path)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}

	if resp.StatusCode() >= 400 {
		var apiErr APIError
		if err := json.Unmarshal(resp.Body(), &apiErr); err != nil {
			return NewAPIError(resp.StatusCode(), string(resp.Body()))
		}
		return &apiErr
	}

	return nil
}

// User methods
func (c *client) GetCurrentUser(ctx context.Context) (*User, error) {
	var user User
	err := c.makeRequest(ctx, http.MethodGet, "/v0/user", nil, &user)
	return &user, err
}

func (c *client) GetUser(ctx context.Context, userID string) (*User, error) {
	var user User
	err := c.makeRequest(ctx, http.MethodGet, fmt.Sprintf("/v0/users/%s", userID), nil, &user)
	return &user, err
}

// Team methods
func (c *client) ListTeams(ctx context.Context, limit, offset int) (*TeamsResponse, error) {
	var resp TeamsResponse
	path := "/v0/teams"
	if limit > 0 || offset > 0 {
		path += "?limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)
	}
	err := c.makeRequest(ctx, http.MethodGet, path, nil, &resp)
	return &resp, err
}

func (c *client) GetTeam(ctx context.Context, teamID string) (*Team, error) {
	var team Team
	err := c.makeRequest(ctx, http.MethodGet, fmt.Sprintf("/v0/teams/%s", teamID), nil, &team)
	return &team, err
}

// Workspace methods
func (c *client) ListWorkspaces(ctx context.Context, limit, offset int) (*WorkspacesResponse, error) {
	var resp WorkspacesResponse
	path := "/v0/workspaces"
	if limit > 0 || offset > 0 {
		path += "?limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)
	}
	err := c.makeRequest(ctx, http.MethodGet, path, nil, &resp)
	return &resp, err
}

func (c *client) GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	var workspace Workspace
	err := c.makeRequest(ctx, http.MethodGet, fmt.Sprintf("/v0/workspaces/%s", workspaceID), nil, &workspace)
	return &workspace, err
}

func (c *client) CreateWorkspace(ctx context.Context, req *CreateWorkspaceRequest) (*Workspace, error) {
	var workspace Workspace
	err := c.makeRequest(ctx, http.MethodPost, "/v0/workspaces", req, &workspace)
	return &workspace, err
}

func (c *client) UpdateWorkspace(ctx context.Context, workspaceID string, req *UpdateWorkspaceRequest) (*Workspace, error) {
	var workspace Workspace
	err := c.makeRequest(ctx, http.MethodPatch, fmt.Sprintf("/v0/workspaces/%s", workspaceID), req, &workspace)
	return &workspace, err
}

func (c *client) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	return c.makeRequest(ctx, http.MethodDelete, fmt.Sprintf("/v0/workspaces/%s", workspaceID), nil, nil)
}

// Collection methods
func (c *client) ListCollections(ctx context.Context, workspaceID string, limit, offset int) (*CollectionsResponse, error) {
	var resp CollectionsResponse
	path := fmt.Sprintf("/v0/workspaces/%s/collections", workspaceID)
	if limit > 0 || offset > 0 {
		path += "?limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)
	}
	err := c.makeRequest(ctx, http.MethodGet, path, nil, &resp)
	return &resp, err
}

func (c *client) GetCollection(ctx context.Context, collectionID string) (*Collection, error) {
	var collection Collection
	err := c.makeRequest(ctx, http.MethodGet, fmt.Sprintf("/v0/collections/%s", collectionID), nil, &collection)
	return &collection, err
}

func (c *client) CreateCollection(ctx context.Context, req *CreateCollectionRequest) (*Collection, error) {
	var collection Collection
	err := c.makeRequest(ctx, http.MethodPost, "/v0/collections", req, &collection)
	return &collection, err
}

func (c *client) UpdateCollection(ctx context.Context, collectionID string, req *UpdateCollectionRequest) (*Collection, error) {
	var collection Collection
	err := c.makeRequest(ctx, http.MethodPatch, fmt.Sprintf("/v0/collections/%s", collectionID), req, &collection)
	return &collection, err
}

func (c *client) DeleteCollection(ctx context.Context, collectionID string) error {
	return c.makeRequest(ctx, http.MethodDelete, fmt.Sprintf("/v0/collections/%s", collectionID), nil, nil)
}

// Item methods
func (c *client) SearchItems(ctx context.Context, req *SearchItemsRequest) (*ItemsResponse, error) {
	var resp ItemsResponse
	err := c.makeRequest(ctx, http.MethodPost, "/v0/items/search", req, &resp)
	return &resp, err
}

func (c *client) ListItems(ctx context.Context, workspaceID string, limit, offset int) (*ItemsResponse, error) {
	var resp ItemsResponse
	path := fmt.Sprintf("/v0/workspaces/%s/items", workspaceID)
	if limit > 0 || offset > 0 {
		path += "?limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)
	}
	err := c.makeRequest(ctx, http.MethodGet, path, nil, &resp)
	return &resp, err
}

func (c *client) GetItem(ctx context.Context, itemID string) (*Item, error) {
	var item Item
	err := c.makeRequest(ctx, http.MethodGet, fmt.Sprintf("/v0/items/%s", itemID), nil, &item)
	return &item, err
}

func (c *client) CreateItem(ctx context.Context, req *CreateItemRequest) (*Item, error) {
	var item Item
	err := c.makeRequest(ctx, http.MethodPost, "/v0/items", req, &item)
	return &item, err
}

func (c *client) UpdateItem(ctx context.Context, itemID string, req *UpdateItemRequest) (*Item, error) {
	var item Item
	err := c.makeRequest(ctx, http.MethodPatch, fmt.Sprintf("/v0/items/%s", itemID), req, &item)
	return &item, err
}

func (c *client) DeleteItem(ctx context.Context, itemID string) error {
	return c.makeRequest(ctx, http.MethodDelete, fmt.Sprintf("/v0/items/%s", itemID), nil, nil)
}

func (c *client) MoveItem(ctx context.Context, itemID, collectionID string) (*Item, error) {
	req := map[string]string{"collectionId": collectionID}
	var item Item
	err := c.makeRequest(ctx, http.MethodPatch, fmt.Sprintf("/v0/items/%s/move", itemID), req, &item)
	return &item, err
}

// File methods
func (c *client) ListFiles(ctx context.Context, workspaceID string, limit, offset int) (*FilesResponse, error) {
	var resp FilesResponse
	path := fmt.Sprintf("/v0/workspaces/%s/files", workspaceID)
	if limit > 0 || offset > 0 {
		path += "?limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)
	}
	err := c.makeRequest(ctx, http.MethodGet, path, nil, &resp)
	return &resp, err
}

func (c *client) GetFile(ctx context.Context, fileID string) (*File, error) {
	var file File
	err := c.makeRequest(ctx, http.MethodGet, fmt.Sprintf("/v0/files/%s", fileID), nil, &file)
	return &file, err
}

func (c *client) UploadFile(ctx context.Context, workspaceID, filename string, data []byte) (*File, error) {
	// Apply rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	var file File
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetFile("file", filename).
		SetFormData(map[string]string{"workspaceId": workspaceID}).
		SetResult(&file).
		Post("/v0/files")

	if err != nil {
		return nil, fmt.Errorf("file upload failed: %w", err)
	}

	if resp.StatusCode() >= 400 {
		var apiErr APIError
		if err := json.Unmarshal(resp.Body(), &apiErr); err != nil {
			return nil, NewAPIError(resp.StatusCode(), string(resp.Body()))
		}
		return nil, &apiErr
	}

	return &file, nil
}

func (c *client) DownloadFile(ctx context.Context, fileID string) ([]byte, error) {
	// Apply rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	resp, err := c.httpClient.R().
		SetContext(ctx).
		Get(fmt.Sprintf("/v0/files/%s/download", fileID))

	if err != nil {
		return nil, fmt.Errorf("file download failed: %w", err)
	}

	if resp.StatusCode() >= 400 {
		var apiErr APIError
		if err := json.Unmarshal(resp.Body(), &apiErr); err != nil {
			return nil, NewAPIError(resp.StatusCode(), string(resp.Body()))
		}
		return nil, &apiErr
	}

	return resp.Body(), nil
}
