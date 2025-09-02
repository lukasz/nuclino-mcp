package tools

import (
	"context"
	"fmt"

	"github.com/lukasz/nuclino-mcp-server/internal/nuclino"
	"github.com/mark3labs/mcp-go/mcp"
)

// GetItemTool implements getting a single item by ID
type GetItemTool struct {
	client nuclino.Client
}

func (t *GetItemTool) Name() string {
	return "nuclino_get_item"
}

func (t *GetItemTool) Description() string {
	return "Get a Nuclino item by ID with full content in Markdown format"
}

func (t *GetItemTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"item_id": StringProperty("The ID of the item to retrieve"),
	}, []string{"item_id"})
}

func (t *GetItemTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	itemID, ok := args["item_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("item_id must be a string"))
	}

	item, err := t.client.GetItem(context.Background(), itemID)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(item)
}

// SearchItemsTool implements searching items with filters
type SearchItemsTool struct {
	client nuclino.Client
}

func (t *SearchItemsTool) Name() string {
	return "nuclino_search_items"
}

func (t *SearchItemsTool) Description() string {
	return "Search Nuclino items with query and optional filters. Returns items with full content."
}

func (t *SearchItemsTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"query":         StringProperty("Search query text"),
		"workspace_id":  StringProperty("Optional workspace ID to limit search scope"),
		"collection_id": StringProperty("Optional collection ID to limit search to specific collection"),
		"limit":         IntProperty("Maximum number of items to return (default: 50)"),
		"offset":        IntProperty("Number of items to skip for pagination (default: 0)"),
	}, []string{})
}

func (t *SearchItemsTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	req := &nuclino.SearchItemsRequest{}

	if query, ok := args["query"].(string); ok {
		req.Query = query
	}

	if workspaceID, ok := args["workspace_id"].(string); ok {
		req.WorkspaceID = workspaceID
	}

	if limit, ok := args["limit"].(float64); ok {
		req.Limit = int(limit)
	} else {
		req.Limit = 50
	}

	if offset, ok := args["offset"].(float64); ok {
		req.Offset = int(offset)
	}

	// Get initial search results
	items, err := t.client.SearchItems(context.Background(), req)
	if err != nil {
		return FormatError(err)
	}

	// If collection_id is specified, filter results by collection
	if collectionID, ok := args["collection_id"].(string); ok && collectionID != "" {
		var filteredItems []nuclino.Item
		for _, item := range items.Results {
			if item.CollectionID == collectionID {
				filteredItems = append(filteredItems, item)
			}
		}

		// Update response with filtered results
		items.Results = filteredItems
		items.Total = len(filteredItems)
	}

	return FormatResult(items)
}

// CreateItemTool implements creating new items
type CreateItemTool struct {
	client nuclino.Client
}

func (t *CreateItemTool) Name() string {
	return "nuclino_create_item"
}

func (t *CreateItemTool) Description() string {
	return "Create a new Nuclino item with title, content (in Markdown), and collection ID"
}

func (t *CreateItemTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"title":         StringProperty("The title of the item"),
		"content":       StringProperty("The content of the item in Markdown format"),
		"collection_id": StringProperty("The ID of the collection to create the item in"),
	}, []string{"title", "collection_id"})
}

func (t *CreateItemTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	req := &nuclino.CreateItemRequest{}

	title, ok := args["title"].(string)
	if !ok {
		return FormatError(fmt.Errorf("title must be a string"))
	}
	req.Title = title

	collectionID, ok := args["collection_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("collection_id must be a string"))
	}
	req.CollectionID = collectionID

	if content, ok := args["content"].(string); ok {
		req.Content = content
	}

	item, err := t.client.CreateItem(context.Background(), req)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(item)
}

// UpdateItemTool implements updating existing items
type UpdateItemTool struct {
	client nuclino.Client
}

func (t *UpdateItemTool) Name() string {
	return "nuclino_update_item"
}

func (t *UpdateItemTool) Description() string {
	return "Update an existing Nuclino item. You can update title, content (Markdown), or collection ID"
}

func (t *UpdateItemTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"item_id":       StringProperty("The ID of the item to update"),
		"title":         StringProperty("New title for the item (optional)"),
		"content":       StringProperty("New content for the item in Markdown format (optional)"),
		"collection_id": StringProperty("New collection ID to move the item to (optional)"),
	}, []string{"item_id"})
}

func (t *UpdateItemTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	itemID, ok := args["item_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("item_id must be a string"))
	}

	req := &nuclino.UpdateItemRequest{}

	if title, ok := args["title"].(string); ok {
		req.Title = &title
	}

	if content, ok := args["content"].(string); ok {
		req.Content = &content
	}

	if collectionID, ok := args["collection_id"].(string); ok {
		req.CollectionID = &collectionID
	}

	item, err := t.client.UpdateItem(context.Background(), itemID, req)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(item)
}

// DeleteItemTool implements soft deleting items
type DeleteItemTool struct {
	client nuclino.Client
}

func (t *DeleteItemTool) Name() string {
	return "nuclino_delete_item"
}

func (t *DeleteItemTool) Description() string {
	return "Delete a Nuclino item (moves to trash). This is a soft delete operation."
}

func (t *DeleteItemTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"item_id": StringProperty("The ID of the item to delete"),
	}, []string{"item_id"})
}

func (t *DeleteItemTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	itemID, ok := args["item_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("item_id must be a string"))
	}

	err := t.client.DeleteItem(context.Background(), itemID)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Item %s has been deleted successfully", itemID),
	})
}

// MoveItemTool implements moving items between collections
type MoveItemTool struct {
	client nuclino.Client
}

func (t *MoveItemTool) Name() string {
	return "nuclino_move_item"
}

func (t *MoveItemTool) Description() string {
	return "Move a Nuclino item from one collection to another"
}

func (t *MoveItemTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"item_id":       StringProperty("The ID of the item to move"),
		"collection_id": StringProperty("The ID of the destination collection"),
	}, []string{"item_id", "collection_id"})
}

func (t *MoveItemTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	itemID, ok := args["item_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("item_id must be a string"))
	}

	collectionID, ok := args["collection_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("collection_id must be a string"))
	}

	item, err := t.client.MoveItem(context.Background(), itemID, collectionID)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(item)
}

// ListItemsTool implements listing items in a workspace
type ListItemsTool struct {
	client nuclino.Client
}

func (t *ListItemsTool) Name() string {
	return "nuclino_list_items"
}

func (t *ListItemsTool) Description() string {
	return "List all items in a Nuclino workspace with pagination support"
}

func (t *ListItemsTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"workspace_id": StringProperty("The ID of the workspace to list items from"),
		"limit":        IntProperty("Maximum number of items to return (default: 50)"),
		"offset":       IntProperty("Number of items to skip for pagination (default: 0)"),
	}, []string{"workspace_id"})
}

func (t *ListItemsTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	workspaceID, ok := args["workspace_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("workspace_id must be a string"))
	}

	limit := 50
	offset := 0

	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}

	items, err := t.client.ListItems(context.Background(), workspaceID, limit, offset)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(items)
}

// ListCollectionItemsTool implements listing items in a specific collection
type ListCollectionItemsTool struct {
	client nuclino.Client
}

func (t *ListCollectionItemsTool) Name() string {
	return "nuclino_list_collection_items"
}

func (t *ListCollectionItemsTool) Description() string {
	return "List all items in a specific Nuclino collection with pagination support"
}

func (t *ListCollectionItemsTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"collection_id": StringProperty("The ID of the collection to list items from"),
		"limit":         IntProperty("Maximum number of items to return (default: 50)"),
		"offset":        IntProperty("Number of items to skip for pagination (default: 0)"),
	}, []string{"collection_id"})
}

func (t *ListCollectionItemsTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	collectionID, ok := args["collection_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("collection_id must be a string"))
	}

	limit := 50
	offset := 0

	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}

	// Get collection info first to find the workspace
	collection, err := t.client.GetCollection(context.Background(), collectionID)
	if err != nil {
		return FormatError(err)
	}

	// Get all items in the workspace and filter by collection
	items, err := t.client.ListItems(context.Background(), collection.WorkspaceID, 1000, 0) // Get more to filter
	if err != nil {
		return FormatError(err)
	}

	// Filter items by collection ID
	var filteredItems []nuclino.Item
	for _, item := range items.Results {
		if item.CollectionID == collectionID {
			filteredItems = append(filteredItems, item)
		}
	}

	// Apply pagination to filtered results
	total := len(filteredItems)
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	if start < end {
		filteredItems = filteredItems[start:end]
	} else {
		filteredItems = []nuclino.Item{}
	}

	response := &nuclino.ItemsResponse{
		Results: filteredItems,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}

	return FormatResult(response)
}
