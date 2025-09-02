package tools

import (
	"context"
	"fmt"

	"github.com/lukasz/nuclino-mcp-server/internal/nuclino"
	"github.com/mark3labs/mcp-go/mcp"
)

// GetWorkspaceOverviewTool provides comprehensive overview of workspace content
type GetWorkspaceOverviewTool struct {
	client nuclino.Client
}

func (t *GetWorkspaceOverviewTool) Name() string {
	return "nuclino_get_workspace_overview"
}

func (t *GetWorkspaceOverviewTool) Description() string {
	return "Get comprehensive overview of a Nuclino workspace including collections, item counts, and recent activity"
}

func (t *GetWorkspaceOverviewTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"workspace_id":   StringProperty("The ID of the workspace to analyze"),
		"include_items":  BoolProperty("Whether to include summary of items (default: false)"),
		"include_recent": BoolProperty("Whether to include recent items (default: false)"),
		"recent_limit":   IntProperty("Number of recent items to include (default: 10)"),
	}, []string{"workspace_id"})
}

func (t *GetWorkspaceOverviewTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	workspaceID, ok := args["workspace_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("workspace_id must be a string"))
	}

	includeItems := false
	if include, ok := args["include_items"].(bool); ok {
		includeItems = include
	}

	includeRecent := false
	if include, ok := args["include_recent"].(bool); ok {
		includeRecent = include
	}

	recentLimit := 10
	if limit, ok := args["recent_limit"].(float64); ok {
		recentLimit = int(limit)
	}

	// Get workspace info
	workspace, err := t.client.GetWorkspace(context.Background(), workspaceID)
	if err != nil {
		return FormatError(err)
	}

	// Get collections in workspace
	collections, err := t.client.ListCollections(context.Background(), workspaceID, 100, 0)
	if err != nil {
		return FormatError(err)
	}

	overview := map[string]interface{}{
		"workspace": workspace,
		"collections": map[string]interface{}{
			"total": collections.Total,
			"items": collections.Results,
		},
	}

	if includeItems {
		// Get items summary
		items, err := t.client.ListItems(context.Background(), workspaceID, 1000, 0) // Get many for counting
		if err != nil {
			return FormatError(err)
		}

		// Count items per collection
		itemCounts := make(map[string]int)
		for _, item := range items.Results {
			itemCounts[item.CollectionID]++
		}

		overview["items_summary"] = map[string]interface{}{
			"total_items":          items.Total,
			"items_per_collection": itemCounts,
		}

		if includeRecent {
			// Get recent items (first N items, assuming they're ordered by update time)
			limit := recentLimit
			if limit > len(items.Results) {
				limit = len(items.Results)
			}

			recentItems := items.Results[:limit]
			overview["recent_items"] = recentItems
		}
	}

	return FormatResult(overview)
}

// SearchWorkspaceContentTool provides advanced search across workspace
type SearchWorkspaceContentTool struct {
	client nuclino.Client
}

func (t *SearchWorkspaceContentTool) Name() string {
	return "nuclino_search_workspace_content"
}

func (t *SearchWorkspaceContentTool) Description() string {
	return "Advanced search within a workspace with content type filtering and aggregated results"
}

func (t *SearchWorkspaceContentTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"workspace_id":        StringProperty("The ID of the workspace to search in"),
		"query":               StringProperty("Search query text"),
		"search_titles":       BoolProperty("Whether to search in titles (default: true)"),
		"search_content":      BoolProperty("Whether to search in content (default: true)"),
		"group_by_collection": BoolProperty("Whether to group results by collection (default: false)"),
		"limit":               IntProperty("Maximum number of items to return (default: 50)"),
	}, []string{"workspace_id", "query"})
}

func (t *SearchWorkspaceContentTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	workspaceID, ok := args["workspace_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("workspace_id must be a string"))
	}

	query, ok := args["query"].(string)
	if !ok {
		return FormatError(fmt.Errorf("query must be a string"))
	}

	searchTitles := true
	if search, ok := args["search_titles"].(bool); ok {
		searchTitles = search
	}

	searchContent := true
	if search, ok := args["search_content"].(bool); ok {
		searchContent = search
	}

	groupByCollection := false
	if group, ok := args["group_by_collection"].(bool); ok {
		groupByCollection = group
	}

	limit := 50
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	// Search items in workspace
	searchReq := &nuclino.SearchItemsRequest{
		Query:       query,
		WorkspaceID: workspaceID,
		Limit:       limit * 2, // Get more for filtering
		Offset:      0,
	}

	items, err := t.client.SearchItems(context.Background(), searchReq)
	if err != nil {
		return FormatError(err)
	}

	// Filter based on search preferences
	var filteredItems []nuclino.Item
	for _, item := range items.Results {
		matches := false

		if searchTitles && containsIgnoreCase(item.Title, query) {
			matches = true
		}

		if searchContent && containsIgnoreCase(item.Content, query) {
			matches = true
		}

		if matches {
			filteredItems = append(filteredItems, item)
		}

		if len(filteredItems) >= limit {
			break
		}
	}

	result := map[string]interface{}{
		"query":        query,
		"workspace_id": workspaceID,
		"total_found":  len(filteredItems),
		"items":        filteredItems,
	}

	if groupByCollection {
		// Group results by collection
		groupedResults := make(map[string][]nuclino.Item)
		for _, item := range filteredItems {
			groupedResults[item.CollectionID] = append(groupedResults[item.CollectionID], item)
		}

		result["grouped_by_collection"] = groupedResults
	}

	return FormatResult(result)
}

// Helper function for case-insensitive string contains
func containsIgnoreCase(str, substr string) bool {
	str = toLower(str)
	substr = toLower(substr)
	return contains(str, substr)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i, b := range []byte(s) {
		if b >= 'A' && b <= 'Z' {
			result[i] = b + 32
		} else {
			result[i] = b
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || indexOfSubstring(s, substr) >= 0)
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
