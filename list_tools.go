package main

import (
	"fmt"

	"github.com/lukasz/nuclino-mcp-server/internal/nuclino"
	"github.com/lukasz/nuclino-mcp-server/internal/tools"
)

func main() {
	// Create dummy client
	client := nuclino.NewClient("dummy")
	
	// Create registry
	registry := tools.NewRegistry(client)
	
	// List all tools
	toolsList := registry.ListTools()
	
	fmt.Printf("📋 Available MCP tools: %d\n", len(toolsList))
	fmt.Println("=" + string(make([]byte, 50)))
	
	for i, tool := range toolsList {
		fmt.Printf("%d. %s\n", i+1, tool.Name)
		fmt.Printf("   Description: %s\n", tool.Description)
		
		// Show required parameters
		if tool.InputSchema.Properties != nil {
			fmt.Print("   Required: ")
			// This is tricky because we don't have access to the original required array
			// But we can check if certain problematic params exist
			if _, hasCollectionID := tool.InputSchema.Properties["collection_id"]; hasCollectionID {
				fmt.Print("❌ collection_id (PROBLEMATIC!)")
			}
			if _, hasWorkspaceID := tool.InputSchema.Properties["workspace_id"]; hasWorkspaceID {
				fmt.Print("✅ workspace_id")
			}
			fmt.Println()
		}
		fmt.Println()
	}
	
	// Look specifically for create_item tool
	fmt.Println("\n🔍 Looking for create_item tool:")
	for _, tool := range toolsList {
		if tool.Name == "nuclino_create_item" {
			fmt.Printf("✅ Found: %s\n", tool.Name)
			fmt.Printf("   Description: %s\n", tool.Description)
			fmt.Printf("   Properties: %v\n", tool.InputSchema.Properties)
			
			if props := tool.InputSchema.Properties; props != nil {
				if _, hasCollection := props["collection_id"]; hasCollection {
					fmt.Println("   ❌ HAS collection_id parameter!")
				}
				if _, hasWorkspace := props["workspace_id"]; hasWorkspace {
					fmt.Println("   ✅ HAS workspace_id parameter!")
				}
			}
		}
	}
}