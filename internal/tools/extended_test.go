package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lukasz/nuclino-mcp-server/internal/nuclino"
)

// Test ListItemsTool
func TestListItemsTool_Execute_Success(t *testing.T) {
	mockClient := new(MockClient)
	tool := &ListItemsTool{client: mockClient}

	expectedResponse := &nuclino.ItemsResponse{
		Results: []nuclino.Item{
			{ID: "item-1", Title: "Item 1", WorkspaceID: "workspace-123"},
			{ID: "item-2", Title: "Item 2", WorkspaceID: "workspace-123"},
		},
		Total: 2, Limit: 50, Offset: 0,
	}

	mockClient.On("ListItems", mock.Anything, "workspace-123", 50, 0).Return(expectedResponse, nil)

	args := map[string]interface{}{
		"workspace_id": "workspace-123",
	}

	result, err := tool.Execute(args)

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Len(t, result.Content, 1)

	mockClient.AssertExpectations(t)
}

func TestListItemsTool_Execute_InvalidArgs(t *testing.T) {
	mockClient := new(MockClient)
	tool := &ListItemsTool{client: mockClient}

	args := map[string]interface{}{
		"workspace_id": 123, // Should be string
	}

	result, err := tool.Execute(args)

	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Len(t, result.Content, 1)

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			assert.Contains(t, textContent.Text, "workspace_id must be a string")
		}
	}
}

// Test ListCollectionItemsTool
func TestListCollectionItemsTool_Execute_Success(t *testing.T) {
	mockClient := new(MockClient)
	tool := &ListCollectionItemsTool{client: mockClient}

	collection := &nuclino.Collection{
		ID:          "collection-123",
		Title:       "Test Collection",
		WorkspaceID: "workspace-456",
	}

	workspaceItems := &nuclino.ItemsResponse{
		Results: []nuclino.Item{
			{ID: "item-1", Title: "Item 1", CollectionID: "collection-123", WorkspaceID: "workspace-456"},
			{ID: "item-2", Title: "Item 2", CollectionID: "collection-456", WorkspaceID: "workspace-456"}, // Different collection
			{ID: "item-3", Title: "Item 3", CollectionID: "collection-123", WorkspaceID: "workspace-456"},
		},
		Total: 3, Limit: 1000, Offset: 0,
	}

	mockClient.On("GetCollection", mock.Anything, "collection-123").Return(collection, nil)
	mockClient.On("ListItems", mock.Anything, "workspace-456", 1000, 0).Return(workspaceItems, nil)

	args := map[string]interface{}{
		"collection_id": "collection-123",
	}

	result, err := tool.Execute(args)

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Len(t, result.Content, 1)

	mockClient.AssertExpectations(t)
}

// Test SearchItemsTool with collection filtering
func TestSearchItemsTool_Execute_WithCollectionFilter(t *testing.T) {
	mockClient := new(MockClient)
	tool := &SearchItemsTool{client: mockClient}

	searchResponse := &nuclino.ItemsResponse{
		Results: []nuclino.Item{
			{ID: "item-1", Title: "Test Item 1", CollectionID: "collection-123"},
			{ID: "item-2", Title: "Test Item 2", CollectionID: "collection-456"},
			{ID: "item-3", Title: "Test Item 3", CollectionID: "collection-123"},
		},
		Total: 3, Limit: 50, Offset: 0,
	}

	mockClient.On("SearchItems", mock.Anything, mock.MatchedBy(func(req *nuclino.SearchItemsRequest) bool {
		return req.Query == "test" && req.Limit == 50
	})).Return(searchResponse, nil)

	args := map[string]interface{}{
		"query":         "test",
		"collection_id": "collection-123", // Filter by this collection
	}

	result, err := tool.Execute(args)

	assert.NoError(t, err)
	assert.False(t, result.IsError)

	mockClient.AssertExpectations(t)
}

// Test GetWorkspaceOverviewTool
func TestGetWorkspaceOverviewTool_Execute_Success(t *testing.T) {
	mockClient := new(MockClient)
	tool := &GetWorkspaceOverviewTool{client: mockClient}

	workspace := &nuclino.Workspace{
		ID:     "workspace-123",
		Name:   "Test Workspace",
		TeamID: "team-456",
	}

	collections := &nuclino.CollectionsResponse{
		Results: []nuclino.Collection{
			{ID: "collection-1", Title: "Collection 1", WorkspaceID: "workspace-123"},
			{ID: "collection-2", Title: "Collection 2", WorkspaceID: "workspace-123"},
		},
		Total: 2, Limit: 100, Offset: 0,
	}

	mockClient.On("GetWorkspace", mock.Anything, "workspace-123").Return(workspace, nil)
	mockClient.On("ListCollections", mock.Anything, "workspace-123", 100, 0).Return(collections, nil)

	args := map[string]interface{}{
		"workspace_id": "workspace-123",
	}

	result, err := tool.Execute(args)

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Len(t, result.Content, 1)

	mockClient.AssertExpectations(t)
}

// Test GetCollectionOverviewTool
func TestGetCollectionOverviewTool_Execute_Success(t *testing.T) {
	mockClient := new(MockClient)
	tool := &GetCollectionOverviewTool{client: mockClient}

	collection := &nuclino.Collection{
		ID:          "collection-123",
		Title:       "Test Collection",
		WorkspaceID: "workspace-456",
	}

	items := &nuclino.ItemsResponse{
		Results: []nuclino.Item{
			{ID: "item-1", Title: "Item 1", CollectionID: "collection-123", Content: "Some content"},
			{ID: "item-2", Title: "Item 2", CollectionID: "collection-123", Content: "More content"},
			{ID: "item-3", Title: "Item 3", CollectionID: "collection-456", Content: "Different collection"},
		},
		Total: 3, Limit: 1000, Offset: 0,
	}

	mockClient.On("GetCollection", mock.Anything, "collection-123").Return(collection, nil)
	mockClient.On("ListItems", mock.Anything, "workspace-456", 1000, 0).Return(items, nil)

	args := map[string]interface{}{
		"collection_id":      "collection-123",
		"include_statistics": true,
		"include_recent":     true,
	}

	result, err := tool.Execute(args)

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Len(t, result.Content, 1)

	mockClient.AssertExpectations(t)
}

// Test BulkOperationsTool dry run
func TestBulkOperationsTool_Execute_DryRun(t *testing.T) {
	mockClient := new(MockClient)
	tool := &BulkOperationsTool{client: mockClient}

	collection := &nuclino.Collection{
		ID:          "source-collection",
		Title:       "Source Collection",
		WorkspaceID: "workspace-123",
	}

	items := &nuclino.ItemsResponse{
		Results: []nuclino.Item{
			{ID: "item-1", Title: "Item 1", CollectionID: "source-collection"},
			{ID: "item-2", Title: "Item 2", CollectionID: "source-collection"},
		},
		Total: 2, Limit: 1000, Offset: 0,
	}

	mockClient.On("GetCollection", mock.Anything, "source-collection").Return(collection, nil)
	mockClient.On("ListItems", mock.Anything, "workspace-123", 1000, 0).Return(items, nil)

	args := map[string]interface{}{
		"operation":         "move",
		"source_collection": "source-collection",
		"target_collection": "target-collection",
		"dry_run":           true,
	}

	result, err := tool.Execute(args)

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Len(t, result.Content, 1)

	mockClient.AssertExpectations(t)
}

// Test tool name and description methods
func TestExtendedTools_Metadata(t *testing.T) {
	mockClient := new(MockClient)

	tools := []Tool{
		&ListItemsTool{client: mockClient},
		&ListCollectionItemsTool{client: mockClient},
		&GetWorkspaceOverviewTool{client: mockClient},
		&SearchWorkspaceContentTool{client: mockClient},
		&GetCollectionOverviewTool{client: mockClient},
		&OrganizeCollectionTool{client: mockClient},
		&BulkOperationsTool{client: mockClient},
	}

	for _, tool := range tools {
		assert.NotEmpty(t, tool.Name(), "Tool should have a name")
		assert.NotEmpty(t, tool.Description(), "Tool should have a description")
		assert.NotNil(t, tool.InputSchema(), "Tool should have an input schema")
	}
}

// Test registry with all tools
func TestRegistry_WithExtendedTools(t *testing.T) {
	mockClient := new(MockClient)
	registry := NewRegistry(mockClient)

	tools := registry.ListTools()

	// Should have all basic + extended tools
	assert.Greater(t, len(tools), 22, "Should have more than 22 tools including extended ones")

	// Check for presence of extended tools
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	extendedTools := []string{
		"nuclino_list_items",
		"nuclino_list_collection_items",
		"nuclino_get_workspace_overview",
		"nuclino_search_workspace_content",
		"nuclino_get_collection_overview",
		"nuclino_organize_collection",
		"nuclino_bulk_collection_operations",
	}

	for _, toolName := range extendedTools {
		assert.True(t, toolNames[toolName], "Should contain extended tool: %s", toolName)
	}
}
