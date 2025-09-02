package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lukasz/nuclino-mcp-server/internal/nuclino"
)

// Integration tests for extended Collections, Workspaces and Search functionality
// These tests verify that tools work together correctly and handle complex scenarios

func TestWorkspaceAndCollectionIntegration(t *testing.T) {
	mockClient := new(MockClient)
	registry := NewRegistry(mockClient)

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

	items := &nuclino.ItemsResponse{
		Results: []nuclino.Item{
			{ID: "item-1", Title: "Item 1", CollectionID: "collection-1", WorkspaceID: "workspace-123", Content: "First item content"},
			{ID: "item-2", Title: "Item 2", CollectionID: "collection-1", WorkspaceID: "workspace-123", Content: "Second item content"},
			{ID: "item-3", Title: "Item 3", CollectionID: "collection-2", WorkspaceID: "workspace-123", Content: "Third item content"},
		},
		Total: 3, Limit: 1000, Offset: 0,
	}

	mockClient.On("GetWorkspace", mock.Anything, "workspace-123").Return(workspace, nil)
	mockClient.On("ListCollections", mock.Anything, "workspace-123", 100, 0).Return(collections, nil)
	mockClient.On("ListItems", mock.Anything, "workspace-123", 1000, 0).Return(items, nil)

	// Test workspace overview with items
	overviewArgs := map[string]interface{}{
		"workspace_id":   "workspace-123",
		"include_items":  true,
		"include_recent": true,
		"recent_limit":   2.0,
	}

	result, err := registry.CallTool("nuclino_get_workspace_overview", overviewArgs)

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Len(t, result.Content, 1)

	mockClient.AssertExpectations(t)
}

func TestCollectionOrganizationWorkflow(t *testing.T) {
	mockClient := new(MockClient)
	registry := NewRegistry(mockClient)

	collection := &nuclino.Collection{
		ID:          "collection-123",
		Title:       "Test Collection",
		WorkspaceID: "workspace-456",
	}

	items := &nuclino.ItemsResponse{
		Results: []nuclino.Item{
			{ID: "item-1", Title: "Project Alpha", CollectionID: "collection-123", Content: "Alpha project documentation"},
			{ID: "item-2", Title: "Project Beta", CollectionID: "collection-123", Content: "Beta project notes"},
			{ID: "item-3", Title: "Meeting Notes", CollectionID: "collection-123", Content: "Weekly team meeting"},
			{ID: "item-4", Title: "Alpha Update", CollectionID: "collection-123", Content: "Project Alpha progress update"},
			{ID: "item-5", Title: "", CollectionID: "collection-123", Content: ""}, // Empty item
		},
		Total: 5, Limit: 1000, Offset: 0,
	}

	mockClient.On("GetCollection", mock.Anything, "collection-123").Return(collection, nil)
	mockClient.On("ListItems", mock.Anything, "workspace-456", 1000, 0).Return(items, nil)

	// Test collection overview
	overviewArgs := map[string]interface{}{
		"collection_id":      "collection-123",
		"include_statistics": true,
		"include_recent":     true,
		"recent_limit":       3.0,
	}

	overviewResult, err := registry.CallTool("nuclino_get_collection_overview", overviewArgs)

	assert.NoError(t, err)
	assert.False(t, overviewResult.IsError)

	// Test organization suggestions
	organizeArgs := map[string]interface{}{
		"collection_id":     "collection-123",
		"suggest_tags":      true,
		"find_duplicates":   true,
		"analyze_structure": true,
	}

	organizeResult, err := registry.CallTool("nuclino_organize_collection", organizeArgs)

	assert.NoError(t, err)
	assert.False(t, organizeResult.IsError)

	mockClient.AssertExpectations(t)
}

func TestSearchAndFilterWorkflow(t *testing.T) {
	mockClient := new(MockClient)
	registry := NewRegistry(mockClient)

	searchResponse1 := &nuclino.ItemsResponse{
		Results: []nuclino.Item{
			{ID: "item-1", Title: "API Documentation", CollectionID: "collection-dev", WorkspaceID: "workspace-123", Content: "REST API docs"},
			{ID: "item-2", Title: "User Guide", CollectionID: "collection-docs", WorkspaceID: "workspace-123", Content: "User documentation"},
			{ID: "item-3", Title: "API Testing", CollectionID: "collection-dev", WorkspaceID: "workspace-123", Content: "API test cases"},
		},
		Total: 3, Limit: 100, Offset: 0,
	}

	searchResponse2 := &nuclino.ItemsResponse{
		Results: []nuclino.Item{
			{ID: "item-1", Title: "API Documentation", CollectionID: "collection-dev", WorkspaceID: "workspace-123", Content: "REST API docs"},
			{ID: "item-3", Title: "API Testing", CollectionID: "collection-dev", WorkspaceID: "workspace-123", Content: "API test cases"},
		},
		Total: 2, Limit: 20, Offset: 0,
	}

	// Mock for workspace content search
	mockClient.On("SearchItems", mock.Anything, mock.MatchedBy(func(req *nuclino.SearchItemsRequest) bool {
		return req.Query == "API" && req.WorkspaceID == "workspace-123" && req.Limit == 100
	})).Return(searchResponse1, nil).Once()

	// Mock for collection filtered search
	mockClient.On("SearchItems", mock.Anything, mock.MatchedBy(func(req *nuclino.SearchItemsRequest) bool {
		return req.Query == "API" && req.Limit == 10
	})).Return(searchResponse2, nil).Once()

	// Test workspace content search
	searchArgs := map[string]interface{}{
		"workspace_id":        "workspace-123",
		"query":               "API",
		"search_titles":       true,
		"search_content":      true,
		"group_by_collection": true,
		"limit":               50.0,
	}

	searchResult, err := registry.CallTool("nuclino_search_workspace_content", searchArgs)

	assert.NoError(t, err)
	assert.False(t, searchResult.IsError)

	// Test search with collection filtering
	filterArgs := map[string]interface{}{
		"query":         "API",
		"collection_id": "collection-dev",
		"limit":         10.0,
	}

	filterResult, err := registry.CallTool("nuclino_search_items", filterArgs)

	assert.NoError(t, err)
	assert.False(t, filterResult.IsError)

	mockClient.AssertExpectations(t)
}

func TestBulkOperationsWorkflow(t *testing.T) {
	mockClient := new(MockClient)
	registry := NewRegistry(mockClient)

	sourceCollection := &nuclino.Collection{
		ID:          "source-collection",
		Title:       "Source Collection",
		WorkspaceID: "workspace-123",
	}

	items := &nuclino.ItemsResponse{
		Results: []nuclino.Item{
			{ID: "item-1", Title: "Old Document 1", CollectionID: "source-collection", Content: "Legacy content"},
			{ID: "item-2", Title: "Old Document 2", CollectionID: "source-collection", Content: "More legacy content"},
			{ID: "item-3", Title: "Current Doc", CollectionID: "source-collection", Content: "Current content"},
		},
		Total: 3, Limit: 1000, Offset: 0,
	}

	mockClient.On("GetCollection", mock.Anything, "source-collection").Return(sourceCollection, nil)
	mockClient.On("ListItems", mock.Anything, "workspace-123", 1000, 0).Return(items, nil)

	// Test dry run bulk move operation
	dryRunArgs := map[string]interface{}{
		"operation":         "move",
		"source_collection": "source-collection",
		"target_collection": "target-collection",
		"filter_query":      "Old",
		"dry_run":           true,
	}

	dryRunResult, err := registry.CallTool("nuclino_bulk_collection_operations", dryRunArgs)

	assert.NoError(t, err)
	assert.False(t, dryRunResult.IsError)

	// Test organize operation
	organizeArgs := map[string]interface{}{
		"operation":         "organize",
		"source_collection": "source-collection",
		"dry_run":           true,
	}

	organizeResult, err := registry.CallTool("nuclino_bulk_collection_operations", organizeArgs)

	assert.NoError(t, err)
	assert.False(t, organizeResult.IsError)

	mockClient.AssertExpectations(t)
}

func TestListingToolsIntegration(t *testing.T) {
	mockClient := new(MockClient)
	registry := NewRegistry(mockClient)

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
		Total: 3, Limit: 50, Offset: 0,
	}

	mockClient.On("ListItems", mock.Anything, "workspace-456", 50, 0).Return(workspaceItems, nil)
	mockClient.On("GetCollection", mock.Anything, "collection-123").Return(collection, nil)
	mockClient.On("ListItems", mock.Anything, "workspace-456", 1000, 0).Return(workspaceItems, nil)

	// Test workspace listing
	workspaceArgs := map[string]interface{}{
		"workspace_id": "workspace-456",
		"limit":        50.0,
		"offset":       0.0,
	}

	workspaceResult, err := registry.CallTool("nuclino_list_items", workspaceArgs)

	assert.NoError(t, err)
	assert.False(t, workspaceResult.IsError)

	// Test collection-specific listing
	collectionArgs := map[string]interface{}{
		"collection_id": "collection-123",
		"limit":         50.0,
		"offset":        0.0,
	}

	collectionResult, err := registry.CallTool("nuclino_list_collection_items", collectionArgs)

	assert.NoError(t, err)
	assert.False(t, collectionResult.IsError)

	mockClient.AssertExpectations(t)
}

func TestErrorHandlingInExtendedTools(t *testing.T) {
	mockClient := new(MockClient)
	registry := NewRegistry(mockClient)

	// Test invalid arguments for extended tools
	tests := []struct {
		toolName string
		args     map[string]interface{}
		expected string
	}{
		{
			"nuclino_get_workspace_overview",
			map[string]interface{}{"workspace_id": 123},
			"workspace_id must be a string",
		},
		{
			"nuclino_get_collection_overview",
			map[string]interface{}{"collection_id": 123},
			"collection_id must be a string",
		},
		{
			"nuclino_search_workspace_content",
			map[string]interface{}{"workspace_id": "ws-123"},
			"query must be a string",
		},
		{
			"nuclino_bulk_collection_operations",
			map[string]interface{}{"operation": 123},
			"operation must be a string",
		},
	}

	for _, test := range tests {
		result, err := registry.CallTool(test.toolName, test.args)

		assert.NoError(t, err, "Tool %s should not return error", test.toolName)
		assert.True(t, result.IsError, "Tool %s should set IsError=true", test.toolName)
		assert.Len(t, result.Content, 1, "Tool %s should have error content", test.toolName)

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				assert.Contains(t, textContent.Text, test.expected, "Tool %s should contain expected error", test.toolName)
			}
		}
	}
}

func TestCompleteExtendedToolsCoverage(t *testing.T) {
	mockClient := new(MockClient)
	registry := NewRegistry(mockClient)

	tools := registry.ListTools()

	// Verify all extended tools are registered
	extendedTools := []string{
		"nuclino_list_items",
		"nuclino_list_collection_items",
		"nuclino_get_workspace_overview",
		"nuclino_search_workspace_content",
		"nuclino_get_collection_overview",
		"nuclino_organize_collection",
		"nuclino_bulk_collection_operations",
	}

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	for _, toolName := range extendedTools {
		assert.True(t, toolNames[toolName], "Extended tool %s should be registered", toolName)
	}

	// Verify total tool count includes extended tools
	assert.GreaterOrEqual(t, len(tools), 29, "Should have at least 29 tools including extended ones")
}
