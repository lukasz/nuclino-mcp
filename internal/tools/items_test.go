package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lukasz/nuclino-mcp-server/internal/nuclino"
)

// MockClient is a mock implementation of nuclino.Client for testing
type MockClient struct {
	mock.Mock
}

func (m *MockClient) GetCurrentUser(ctx context.Context) (*nuclino.User, error) {
	args := m.Called(ctx)
	return args.Get(0).(*nuclino.User), args.Error(1)
}

func (m *MockClient) GetUser(ctx context.Context, userID string) (*nuclino.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*nuclino.User), args.Error(1)
}

func (m *MockClient) ListTeams(ctx context.Context, limit, offset int) (*nuclino.TeamsResponse, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).(*nuclino.TeamsResponse), args.Error(1)
}

func (m *MockClient) GetTeam(ctx context.Context, teamID string) (*nuclino.Team, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).(*nuclino.Team), args.Error(1)
}

func (m *MockClient) ListWorkspaces(ctx context.Context, limit, offset int) (*nuclino.WorkspacesResponse, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).(*nuclino.WorkspacesResponse), args.Error(1)
}

func (m *MockClient) GetWorkspace(ctx context.Context, workspaceID string) (*nuclino.Workspace, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).(*nuclino.Workspace), args.Error(1)
}

func (m *MockClient) CreateWorkspace(ctx context.Context, req *nuclino.CreateWorkspaceRequest) (*nuclino.Workspace, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*nuclino.Workspace), args.Error(1)
}

func (m *MockClient) UpdateWorkspace(ctx context.Context, workspaceID string, req *nuclino.UpdateWorkspaceRequest) (*nuclino.Workspace, error) {
	args := m.Called(ctx, workspaceID, req)
	return args.Get(0).(*nuclino.Workspace), args.Error(1)
}

func (m *MockClient) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	args := m.Called(ctx, workspaceID)
	return args.Error(0)
}

func (m *MockClient) ListCollections(ctx context.Context, workspaceID string, limit, offset int) (*nuclino.CollectionsResponse, error) {
	args := m.Called(ctx, workspaceID, limit, offset)
	return args.Get(0).(*nuclino.CollectionsResponse), args.Error(1)
}

func (m *MockClient) GetCollection(ctx context.Context, collectionID string) (*nuclino.Collection, error) {
	args := m.Called(ctx, collectionID)
	return args.Get(0).(*nuclino.Collection), args.Error(1)
}

func (m *MockClient) CreateCollection(ctx context.Context, req *nuclino.CreateCollectionRequest) (*nuclino.Collection, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*nuclino.Collection), args.Error(1)
}

func (m *MockClient) UpdateCollection(ctx context.Context, collectionID string, req *nuclino.UpdateCollectionRequest) (*nuclino.Collection, error) {
	args := m.Called(ctx, collectionID, req)
	return args.Get(0).(*nuclino.Collection), args.Error(1)
}

func (m *MockClient) DeleteCollection(ctx context.Context, collectionID string) error {
	args := m.Called(ctx, collectionID)
	return args.Error(0)
}

func (m *MockClient) SearchItems(ctx context.Context, req *nuclino.SearchItemsRequest) (*nuclino.ItemsResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*nuclino.ItemsResponse), args.Error(1)
}

func (m *MockClient) ListItems(ctx context.Context, workspaceID string, limit, offset int) (*nuclino.ItemsResponse, error) {
	args := m.Called(ctx, workspaceID, limit, offset)
	return args.Get(0).(*nuclino.ItemsResponse), args.Error(1)
}

func (m *MockClient) GetItem(ctx context.Context, itemID string) (*nuclino.Item, error) {
	args := m.Called(ctx, itemID)
	return args.Get(0).(*nuclino.Item), args.Error(1)
}

func (m *MockClient) CreateItem(ctx context.Context, req *nuclino.CreateItemRequest) (*nuclino.Item, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*nuclino.Item), args.Error(1)
}

func (m *MockClient) UpdateItem(ctx context.Context, itemID string, req *nuclino.UpdateItemRequest) (*nuclino.Item, error) {
	args := m.Called(ctx, itemID, req)
	return args.Get(0).(*nuclino.Item), args.Error(1)
}

func (m *MockClient) DeleteItem(ctx context.Context, itemID string) error {
	args := m.Called(ctx, itemID)
	return args.Error(0)
}

func (m *MockClient) MoveItem(ctx context.Context, itemID, collectionID string) (*nuclino.Item, error) {
	args := m.Called(ctx, itemID, collectionID)
	return args.Get(0).(*nuclino.Item), args.Error(1)
}

func (m *MockClient) ListFiles(ctx context.Context, workspaceID string, limit, offset int) (*nuclino.FilesResponse, error) {
	args := m.Called(ctx, workspaceID, limit, offset)
	return args.Get(0).(*nuclino.FilesResponse), args.Error(1)
}

func (m *MockClient) GetFile(ctx context.Context, fileID string) (*nuclino.File, error) {
	args := m.Called(ctx, fileID)
	return args.Get(0).(*nuclino.File), args.Error(1)
}

func (m *MockClient) UploadFile(ctx context.Context, workspaceID, filename string, data []byte) (*nuclino.File, error) {
	args := m.Called(ctx, workspaceID, filename, data)
	return args.Get(0).(*nuclino.File), args.Error(1)
}

func (m *MockClient) DownloadFile(ctx context.Context, fileID string) ([]byte, error) {
	args := m.Called(ctx, fileID)
	return args.Get(0).([]byte), args.Error(1)
}

// Test GetItemTool
func TestGetItemTool_Execute_Success(t *testing.T) {
	mockClient := new(MockClient)
	tool := &GetItemTool{client: mockClient}

	expectedItem := &nuclino.Item{
		ID:      "item-123",
		Title:   "Test Item",
		Content: "# Test Content\n\nThis is a test item.",
	}

	mockClient.On("GetItem", mock.Anything, "item-123").Return(expectedItem, nil)

	args := map[string]interface{}{
		"item_id": "item-123",
	}

	result, err := tool.Execute(args)

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Len(t, result.Content, 1)

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			assert.Contains(t, textContent.Text, "item-123")
			assert.Contains(t, textContent.Text, "Test Item")
		}
	}

	mockClient.AssertExpectations(t)
}

func TestGetItemTool_Execute_InvalidArgs(t *testing.T) {
	mockClient := new(MockClient)
	tool := &GetItemTool{client: mockClient}

	args := map[string]interface{}{
		"item_id": 123, // Should be string, not int
	}

	result, err := tool.Execute(args)

	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Len(t, result.Content, 1)

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			assert.Contains(t, textContent.Text, "item_id must be a string")
		}
	}
}

func TestGetItemTool_Execute_APIError(t *testing.T) {
	mockClient := new(MockClient)
	tool := &GetItemTool{client: mockClient}

	mockClient.On("GetItem", mock.Anything, "nonexistent").Return((*nuclino.Item)(nil), errors.New("item not found"))

	args := map[string]interface{}{
		"item_id": "nonexistent",
	}

	result, err := tool.Execute(args)

	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Len(t, result.Content, 1)

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			assert.Contains(t, textContent.Text, "item not found")
		}
	}

	mockClient.AssertExpectations(t)
}

// Test SearchItemsTool
func TestSearchItemsTool_Execute_Success(t *testing.T) {
	mockClient := new(MockClient)
	tool := &SearchItemsTool{client: mockClient}

	expectedResponse := &nuclino.ItemsResponse{
		Results: []nuclino.Item{
			{
				ID:    "item-1",
				Title: "First Item",
			},
			{
				ID:    "item-2",
				Title: "Second Item",
			},
		},
		Total:  2,
		Limit:  50,
		Offset: 0,
	}

	mockClient.On("SearchItems", mock.Anything, mock.MatchedBy(func(req *nuclino.SearchItemsRequest) bool {
		return req.Query == "test query" && req.Limit == 50
	})).Return(expectedResponse, nil)

	args := map[string]interface{}{
		"query": "test query",
	}

	result, err := tool.Execute(args)

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Len(t, result.Content, 1)

	mockClient.AssertExpectations(t)
}
