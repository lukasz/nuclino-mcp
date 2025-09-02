package server

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"

	"github.com/lukasz/nuclino-mcp-server/internal/nuclino"
	"github.com/lukasz/nuclino-mcp-server/internal/tools"
)

type NuclinoMCPServer struct {
	nuclinoClient nuclino.Client
	toolRegistry  *tools.Registry
	mcpServer     server.MCPServer
}

func NewNuclinoMCPServer(nuclinoClient nuclino.Client) *NuclinoMCPServer {
	s := &NuclinoMCPServer{
		nuclinoClient: nuclinoClient,
		toolRegistry:  tools.NewRegistry(nuclinoClient),
	}

	// Create MCP server
	s.mcpServer = server.NewDefaultServer("nuclino-mcp-server", "0.1.0")

	// Set up handlers
	s.setupHandlers()

	return s
}

func (s *NuclinoMCPServer) setupHandlers() {
	// Cast to DefaultServer to access handler methods
	if defaultServer, ok := s.mcpServer.(*server.DefaultServer); ok {
		// Set tools handler
		defaultServer.HandleCallTool(func(ctx context.Context, name string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			log.Info().Str("tool", name).Msg("Calling tool")

			result, err := s.toolRegistry.CallTool(name, arguments)
			if err != nil {
				log.Error().Err(err).Str("tool", name).Msg("Tool call failed")
				return &mcp.CallToolResult{
					Content: []interface{}{
						mcp.TextContent{
							Type: "text",
							Text: err.Error(),
						},
					},
					IsError: true,
				}, nil
			}

			return result, nil
		})

		// Set tools list handler
		defaultServer.HandleListTools(func(ctx context.Context, cursor *string) (*mcp.ListToolsResult, error) {
			toolsList := s.toolRegistry.ListTools()
			return &mcp.ListToolsResult{
				Tools: toolsList,
			}, nil
		})
	}
}

func (s *NuclinoMCPServer) Run(ctx context.Context) error {
	log.Info().Msg("Starting Nuclino MCP server")
	return server.ServeStdio(s.mcpServer)
}
