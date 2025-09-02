package tools

import (
	"context"
	"fmt"

	"github.com/lukasz/nuclino-mcp-server/internal/nuclino"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListFilesTool implements listing files in a workspace
type ListFilesTool struct {
	client nuclino.Client
}

func (t *ListFilesTool) Name() string {
	return "nuclino_list_files"
}

func (t *ListFilesTool) Description() string {
	return "List all files in a Nuclino workspace with pagination support"
}

func (t *ListFilesTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"workspace_id": StringProperty("The ID of the workspace to list files from"),
		"limit":        IntProperty("Maximum number of files to return (default: 50)"),
		"offset":       IntProperty("Number of files to skip for pagination (default: 0)"),
	}, []string{"workspace_id"})
}

func (t *ListFilesTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
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

	files, err := t.client.ListFiles(context.Background(), workspaceID, limit, offset)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(files)
}

// GetFileTool implements getting file metadata
type GetFileTool struct {
	client nuclino.Client
}

func (t *GetFileTool) Name() string {
	return "nuclino_get_file"
}

func (t *GetFileTool) Description() string {
	return "Get metadata information about a specific Nuclino file"
}

func (t *GetFileTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"file_id": StringProperty("The ID of the file to retrieve"),
	}, []string{"file_id"})
}

func (t *GetFileTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	fileID, ok := args["file_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("file_id must be a string"))
	}

	file, err := t.client.GetFile(context.Background(), fileID)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(file)
}
