package tools

import (
	"encoding/json"
	"fmt"

	"github.com/lukasz/nuclino-mcp-server/internal/nuclino"
	"github.com/mark3labs/mcp-go/mcp"
)

// Registry manages all available MCP tools
type Registry struct {
	tools  map[string]Tool
	client nuclino.Client
}

// Tool interface defines what each MCP tool must implement
type Tool interface {
	Name() string
	Description() string
	InputSchema() interface{}
	Execute(args map[string]interface{}) (*mcp.CallToolResult, error)
}

// NewRegistry creates a new tools registry
func NewRegistry(client nuclino.Client) *Registry {
	registry := &Registry{
		tools:  make(map[string]Tool),
		client: client,
	}

	// Register all tools
	registry.registerBasicTools()

	return registry
}

func (r *Registry) registerBasicTools() {
	// Register item tools
	r.registerTool(&GetItemTool{client: r.client})
	r.registerTool(&SearchItemsTool{client: r.client})
	r.registerTool(&CreateItemTool{client: r.client})
	r.registerTool(&UpdateItemTool{client: r.client})
	r.registerTool(&DeleteItemTool{client: r.client})
	r.registerTool(&MoveItemTool{client: r.client})

	// Register extended item tools
	r.registerTool(&ListItemsTool{client: r.client})
	r.registerTool(&ListCollectionItemsTool{client: r.client})

	// Register workspace tools
	r.registerTool(&ListWorkspacesTool{client: r.client})
	r.registerTool(&GetWorkspaceTool{client: r.client})
	r.registerTool(&CreateWorkspaceTool{client: r.client})
	r.registerTool(&UpdateWorkspaceTool{client: r.client})
	r.registerTool(&DeleteWorkspaceTool{client: r.client})

	// Register extended workspace tools
	r.registerTool(&GetWorkspaceOverviewTool{client: r.client})
	r.registerTool(&SearchWorkspaceContentTool{client: r.client})

	// Register collection tools
	r.registerTool(&ListCollectionsTool{client: r.client})
	r.registerTool(&GetCollectionTool{client: r.client})
	r.registerTool(&CreateCollectionTool{client: r.client})
	r.registerTool(&UpdateCollectionTool{client: r.client})
	r.registerTool(&DeleteCollectionTool{client: r.client})

	// Register extended collection tools
	r.registerTool(&GetCollectionOverviewTool{client: r.client})
	r.registerTool(&OrganizeCollectionTool{client: r.client})
	r.registerTool(&BulkOperationsTool{client: r.client})

	// Register team and user tools
	r.registerTool(&GetCurrentUserTool{client: r.client})
	r.registerTool(&GetUserTool{client: r.client})
	r.registerTool(&ListTeamsTool{client: r.client})
	r.registerTool(&GetTeamTool{client: r.client})

	// Register file tools
	r.registerTool(&ListFilesTool{client: r.client})
	r.registerTool(&GetFileTool{client: r.client})
}

func (r *Registry) registerTool(tool Tool) {
	r.tools[tool.Name()] = tool
}

// ListTools returns all available tools for MCP tools/list
func (r *Registry) ListTools() []mcp.Tool {
	var tools []mcp.Tool
	for _, tool := range r.tools {
		// Convert interface{} to ToolInputSchema
		schema := mcp.ToolInputSchema{
			Type: "object", // Default type
		}

		if inputSchema := tool.InputSchema(); inputSchema != nil {
			if schemaMap, ok := inputSchema.(map[string]interface{}); ok {
				// Convert properties if they exist
				if props, exists := schemaMap["properties"]; exists {
					if propsMap, ok := props.(map[string]interface{}); ok {
						schema.Properties = make(mcp.ToolInputSchemaProperties)
						for key, value := range propsMap {
							if valueMap, ok := value.(map[string]interface{}); ok {
								schema.Properties[key] = valueMap
							}
						}
					}
				}

				// Set schema type if provided
				if schemaType, exists := schemaMap["type"]; exists {
					if typeStr, ok := schemaType.(string); ok {
						schema.Type = typeStr
					}
				}
			}
		}

		tools = append(tools, mcp.Tool{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: schema,
		})
	}
	return tools
}

// CallTool executes a tool by name
func (r *Registry) CallTool(name string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return tool.Execute(args)
}

// JSONSchema helper for creating input schemas
func JSONSchema(properties map[string]interface{}, required []string) interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

// StringProperty creates a string property for JSON schema
func StringProperty(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"description": description,
	}
}

// IntProperty creates an integer property for JSON schema
func IntProperty(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "integer",
		"description": description,
	}
}

// BoolProperty creates a boolean property for JSON schema
func BoolProperty(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "boolean",
		"description": description,
	}
}

// FormatResult formats a result as JSON string for MCP response
func FormatResult(result interface{}) (*mcp.CallToolResult, error) {
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []interface{}{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error formatting result: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonBytes),
			},
		},
	}, nil
}

// FormatError formats an error for MCP response
func FormatError(err error) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Error: %v", err),
			},
		},
		IsError: true,
	}, nil
}
