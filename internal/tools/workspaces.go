package tools

import (
	"context"
	"fmt"

	"github.com/lukasz/nuclino-mcp-server/internal/nuclino"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListWorkspacesTool implements listing workspaces
type ListWorkspacesTool struct {
	client nuclino.Client
}

func (t *ListWorkspacesTool) Name() string {
	return "nuclino_list_workspaces"
}

func (t *ListWorkspacesTool) Description() string {
	return "List all accessible Nuclino workspaces with pagination support"
}

func (t *ListWorkspacesTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"limit":  IntProperty("Maximum number of workspaces to return (default: 50)"),
		"offset": IntProperty("Number of workspaces to skip for pagination (default: 0)"),
	}, []string{})
}

func (t *ListWorkspacesTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := 50
	offset := 0

	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}

	workspaces, err := t.client.ListWorkspaces(context.Background(), limit, offset)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(workspaces)
}

// GetWorkspaceTool implements getting workspace details
type GetWorkspaceTool struct {
	client nuclino.Client
}

func (t *GetWorkspaceTool) Name() string {
	return "nuclino_get_workspace"
}

func (t *GetWorkspaceTool) Description() string {
	return "Get detailed information about a specific Nuclino workspace"
}

func (t *GetWorkspaceTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"workspace_id": StringProperty("The ID of the workspace to retrieve"),
	}, []string{"workspace_id"})
}

func (t *GetWorkspaceTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	workspaceID, ok := args["workspace_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("workspace_id must be a string"))
	}

	workspace, err := t.client.GetWorkspace(context.Background(), workspaceID)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(workspace)
}

// CreateWorkspaceTool implements creating new workspaces
type CreateWorkspaceTool struct {
	client nuclino.Client
}

func (t *CreateWorkspaceTool) Name() string {
	return "nuclino_create_workspace"
}

func (t *CreateWorkspaceTool) Description() string {
	return "Create a new Nuclino workspace within a team"
}

func (t *CreateWorkspaceTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"name":    StringProperty("The name of the new workspace"),
		"team_id": StringProperty("The ID of the team to create the workspace in"),
	}, []string{"name", "team_id"})
}

func (t *CreateWorkspaceTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	name, ok := args["name"].(string)
	if !ok {
		return FormatError(fmt.Errorf("name must be a string"))
	}

	teamID, ok := args["team_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("team_id must be a string"))
	}

	req := &nuclino.CreateWorkspaceRequest{
		Name:   name,
		TeamID: teamID,
	}

	workspace, err := t.client.CreateWorkspace(context.Background(), req)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(workspace)
}

// UpdateWorkspaceTool implements updating workspaces
type UpdateWorkspaceTool struct {
	client nuclino.Client
}

func (t *UpdateWorkspaceTool) Name() string {
	return "nuclino_update_workspace"
}

func (t *UpdateWorkspaceTool) Description() string {
	return "Update an existing Nuclino workspace (currently supports name changes)"
}

func (t *UpdateWorkspaceTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"workspace_id": StringProperty("The ID of the workspace to update"),
		"name":         StringProperty("The new name for the workspace"),
	}, []string{"workspace_id", "name"})
}

func (t *UpdateWorkspaceTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	workspaceID, ok := args["workspace_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("workspace_id must be a string"))
	}

	name, ok := args["name"].(string)
	if !ok {
		return FormatError(fmt.Errorf("name must be a string"))
	}

	req := &nuclino.UpdateWorkspaceRequest{
		Name: &name,
	}

	workspace, err := t.client.UpdateWorkspace(context.Background(), workspaceID, req)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(workspace)
}

// DeleteWorkspaceTool implements deleting workspaces
type DeleteWorkspaceTool struct {
	client nuclino.Client
}

func (t *DeleteWorkspaceTool) Name() string {
	return "nuclino_delete_workspace"
}

func (t *DeleteWorkspaceTool) Description() string {
	return "Delete a Nuclino workspace. WARNING: This action cannot be undone and will delete all content."
}

func (t *DeleteWorkspaceTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"workspace_id": StringProperty("The ID of the workspace to delete"),
		"confirm":      BoolProperty("Confirmation that you want to delete the workspace (must be true)"),
	}, []string{"workspace_id", "confirm"})
}

func (t *DeleteWorkspaceTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	workspaceID, ok := args["workspace_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("workspace_id must be a string"))
	}

	confirm, ok := args["confirm"].(bool)
	if !ok || !confirm {
		return FormatError(fmt.Errorf("you must set confirm=true to delete a workspace"))
	}

	err := t.client.DeleteWorkspace(context.Background(), workspaceID)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Workspace %s has been deleted successfully", workspaceID),
	})
}
