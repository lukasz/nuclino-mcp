package tools

import (
	"context"
	"fmt"

	"github.com/lukasz/nuclino-mcp-server/internal/nuclino"
	"github.com/mark3labs/mcp-go/mcp"
)

// GetCurrentUserTool implements getting current user information
type GetCurrentUserTool struct {
	client nuclino.Client
}

func (t *GetCurrentUserTool) Name() string {
	return "nuclino_get_current_user"
}

func (t *GetCurrentUserTool) Description() string {
	return "Get information about the currently authenticated Nuclino user"
}

func (t *GetCurrentUserTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{}, []string{})
}

func (t *GetCurrentUserTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	user, err := t.client.GetCurrentUser(context.Background())
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(user)
}

// GetUserTool implements getting user information by ID
type GetUserTool struct {
	client nuclino.Client
}

func (t *GetUserTool) Name() string {
	return "nuclino_get_user"
}

func (t *GetUserTool) Description() string {
	return "Get information about a specific Nuclino user by their ID"
}

func (t *GetUserTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"user_id": StringProperty("The ID of the user to retrieve"),
	}, []string{"user_id"})
}

func (t *GetUserTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	userID, ok := args["user_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("user_id must be a string"))
	}

	user, err := t.client.GetUser(context.Background(), userID)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(user)
}

// ListTeamsTool implements listing teams
type ListTeamsTool struct {
	client nuclino.Client
}

func (t *ListTeamsTool) Name() string {
	return "nuclino_list_teams"
}

func (t *ListTeamsTool) Description() string {
	return "List all accessible Nuclino teams with pagination support"
}

func (t *ListTeamsTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"limit":  IntProperty("Maximum number of teams to return (default: 50)"),
		"offset": IntProperty("Number of teams to skip for pagination (default: 0)"),
	}, []string{})
}

func (t *ListTeamsTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := 50
	offset := 0

	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}

	teams, err := t.client.ListTeams(context.Background(), limit, offset)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(teams)
}

// GetTeamTool implements getting team details
type GetTeamTool struct {
	client nuclino.Client
}

func (t *GetTeamTool) Name() string {
	return "nuclino_get_team"
}

func (t *GetTeamTool) Description() string {
	return "Get detailed information about a specific Nuclino team"
}

func (t *GetTeamTool) InputSchema() interface{} {
	return JSONSchema(map[string]interface{}{
		"team_id": StringProperty("The ID of the team to retrieve"),
	}, []string{"team_id"})
}

func (t *GetTeamTool) Execute(args map[string]interface{}) (*mcp.CallToolResult, error) {
	teamID, ok := args["team_id"].(string)
	if !ok {
		return FormatError(fmt.Errorf("team_id must be a string"))
	}

	team, err := t.client.GetTeam(context.Background(), teamID)
	if err != nil {
		return FormatError(err)
	}

	return FormatResult(team)
}
