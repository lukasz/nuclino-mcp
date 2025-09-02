package tools

import (
	"context"
	"fmt"

	"github.com/lukasz/nuclino-mcp-server/internal/nuclino"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListCollectionsTool implements listing collections in a workspace
type ListCollectionsTool struct {
	client nuclino.Client
}

func (t *ListCollectionsTool) Name() string {
	return "nuclino_list_collections"
}

func (t *ListCollectionsTool) Description() string {
	return "List all collections in a Nuclino workspace with pagination support"
}

func (t *ListCollectionsTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"workspace_id": StringProperty("The ID of the workspace to list collections from"),
		"limit":        IntProperty("Maximum number of collections to return (default: 50)"),
		"offset":       IntProperty("Number of collections to skip for pagination (default: 0)"),
	}, []string{"workspace_id"})
}

func (t *ListCollectionsTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
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

	collections, err := t.client.ListCollections(context.Background(), workspaceID, limit, offset)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(collections)
}

// GetCollectionTool implements getting collection details
type GetCollectionTool struct {
	client nuclino.Client
}

func (t *GetCollectionTool) Name() string {
	return "nuclino_get_collection"
}

func (t *GetCollectionTool) Description() string {
	return "Get detailed information about a specific Nuclino collection"
}

func (t *GetCollectionTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"collection_id": StringProperty("The ID of the collection to retrieve"),
	}, []string{"collection_id"})
}

func (t *GetCollectionTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	collectionID, ok := args["collection_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("collection_id must be a string"))
	}

	collection, err := t.client.GetCollection(context.Background(), collectionID)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(collection)
}

// CreateCollectionTool implements creating new collections
type CreateCollectionTool struct {
	client nuclino.Client
}

func (t *CreateCollectionTool) Name() string {
	return "nuclino_create_collection"
}

func (t *CreateCollectionTool) Description() string {
	return "Create a new Nuclino collection within a workspace"
}

func (t *CreateCollectionTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"title":        StringProperty("The title of the new collection"),
		"workspace_id": StringProperty("The ID of the workspace to create the collection in"),
	}, []string{"title", "workspace_id"})
}

func (t *CreateCollectionTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	title, ok := args["title"].(string)
	if !ok {
		return FormatError(fmt.Errorf("title must be a string"))
	}

	workspaceID, ok := args["workspace_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("workspace_id must be a string"))
	}

	req := &nuclino.CreateCollectionRequest{
		Title:       title,
		WorkspaceID: workspaceID,
	}

	collection, err := t.client.CreateCollection(context.Background(), req)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(collection)
}

// UpdateCollectionTool implements updating collections
type UpdateCollectionTool struct {
	client nuclino.Client
}

func (t *UpdateCollectionTool) Name() string {
	return "nuclino_update_collection"
}

func (t *UpdateCollectionTool) Description() string {
	return "Update an existing Nuclino collection (currently supports title changes)"
}

func (t *UpdateCollectionTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"collection_id": StringProperty("The ID of the collection to update"),
		"title":         StringProperty("The new title for the collection"),
	}, []string{"collection_id", "title"})
}

func (t *UpdateCollectionTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	collectionID, ok := args["collection_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("collection_id must be a string"))
	}

	title, ok := args["title"].(string)
	if !ok {
		return FormatError(fmt.Errorf("title must be a string"))
	}

	req := &nuclino.UpdateCollectionRequest{
		Title: &title,
	}

	collection, err := t.client.UpdateCollection(context.Background(), collectionID, req)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(collection)
}

// DeleteCollectionTool implements deleting collections
type DeleteCollectionTool struct {
	client nuclino.Client
}

func (t *DeleteCollectionTool) Name() string {
	return "nuclino_delete_collection"
}

func (t *DeleteCollectionTool) Description() string {
	return "Delete a Nuclino collection. WARNING: This will also delete all items in the collection."
}

func (t *DeleteCollectionTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"collection_id": StringProperty("The ID of the collection to delete"),
		"confirm":       BoolProperty("Confirmation that you want to delete the collection (must be true)"),
	}, []string{"collection_id", "confirm"})
}

func (t *DeleteCollectionTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	collectionID, ok := args["collection_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("collection_id must be a string"))
	}

	confirm, ok := args["confirm"].(bool)
	if !ok || !confirm {
		return FormatError(fmt.Errorf("you must set confirm=true to delete a collection"))
	}

	err := t.client.DeleteCollection(context.Background(), collectionID)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Collection %s has been deleted successfully", collectionID),
	})
}
