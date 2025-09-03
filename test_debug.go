package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/lukasz/nuclino-mcp-server/internal/nuclino"
)

func main() {
	// Setup debug logging
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	apiKey := os.Getenv("NUCLINO_API_KEY")
	if apiKey == "" {
		log.Fatal().Msg("NUCLINO_API_KEY environment variable is required")
	}

	// Create client
	client := nuclino.NewClient(apiKey)
	ctx := context.Background()

	fmt.Println("🔍 Testing Nuclino API endpoints")
	fmt.Println("================================")

	// Test 1: List workspaces
	fmt.Println("\n1. Testing ListWorkspaces...")
	workspaces, err := client.ListWorkspaces(ctx, 10, 0)
	if err != nil {
		log.Error().Err(err).Msg("ListWorkspaces failed")
	} else {
		fmt.Printf("✅ Found %d workspaces\n", len(workspaces.Results))
		if len(workspaces.Results) > 0 {
			workspace := workspaces.Results[0]
			fmt.Printf("   First workspace: %s (ID: %s)\n", workspace.Name, workspace.ID)

			// Test 2: Get workspace details
			fmt.Println("\n2. Testing GetWorkspace...")
			ws, err := client.GetWorkspace(ctx, workspace.ID)
			if err != nil {
				log.Error().Err(err).Msg("GetWorkspace failed")
			} else {
				fmt.Printf("✅ Workspace details: %s (Team ID: %s)\n", ws.Name, ws.TeamID)
			}

			// Test 3: List items in workspace
			fmt.Println("\n3. Testing ListItems...")
			items, err := client.ListItems(ctx, workspace.ID, 10, 0)
			if err != nil {
				log.Error().Err(err).Msg("ListItems failed")
			} else {
				fmt.Printf("✅ Found %d items in workspace\n", len(items.Results))
				if len(items.Results) > 0 {
					item := items.Results[0]
					fmt.Printf("   First item: %s (ID: %s)\n", item.Title, item.ID)

					// Test 4: Get item details
					fmt.Println("\n4. Testing GetItem...")
					itemDetail, err := client.GetItem(ctx, item.ID)
					if err != nil {
						log.Error().Err(err).Msg("GetItem failed")
					} else {
						fmt.Printf("✅ Item details: %s\n", itemDetail.Title)
						fmt.Printf("   Content length: %d characters\n", len(itemDetail.Content))
					}
				}

				// Test 5: Search items
				fmt.Println("\n5. Testing SearchItems...")
				searchReq := &nuclino.SearchItemsRequest{
					Query:       "test",
					WorkspaceID: workspace.ID,
					Limit:       5,
				}
				searchResults, err := client.SearchItems(ctx, searchReq)
				if err != nil {
					log.Error().Err(err).Msg("SearchItems failed")
				} else {
					fmt.Printf("✅ Search found %d items\n", len(searchResults.Results))
				}
			}

			// Test 6: Create item (test the new workspaceId parameter)
			fmt.Println("\n6. Testing CreateItem...")
			createReq := &nuclino.CreateItemRequest{
				Title:       "Test Item " + time.Now().Format("2006-01-02 15:04:05"),
				Content:     "This is a test item created by the debug script.",
				WorkspaceID: workspace.ID,
			}
			newItem, err := client.CreateItem(ctx, createReq)
			if err != nil {
				log.Error().Err(err).Msg("CreateItem failed")
			} else {
				fmt.Printf("✅ Created new item: %s (ID: %s)\n", newItem.Title, newItem.ID)

				// Test 7: Update the created item
				fmt.Println("\n7. Testing UpdateItem...")
				updateContent := "Updated content at " + time.Now().Format("2006-01-02 15:04:05")
				updateReq := &nuclino.UpdateItemRequest{
					Content: &updateContent,
				}
				updatedItem, err := client.UpdateItem(ctx, newItem.ID, updateReq)
				if err != nil {
					log.Error().Err(err).Msg("UpdateItem failed")
				} else {
					fmt.Printf("✅ Updated item: %s\n", updatedItem.Title)
				}

				// Test 8: Delete the created item
				fmt.Println("\n8. Testing DeleteItem...")
				err = client.DeleteItem(ctx, newItem.ID)
				if err != nil {
					log.Error().Err(err).Msg("DeleteItem failed")
				} else {
					fmt.Println("✅ Item deleted successfully")
				}
			}

			// Test 9: List collections (might fail based on API structure)
			fmt.Println("\n9. Testing ListCollections...")
			collections, err := client.ListCollections(ctx, workspace.ID, 10, 0)
			if err != nil {
				log.Error().Err(err).Msg("ListCollections failed")
			} else {
				fmt.Printf("✅ Found %d collections\n", len(collections.Results))
			}
		}
	}

	// Test 10: List teams
	fmt.Println("\n10. Testing ListTeams...")
	teams, err := client.ListTeams(ctx, 10, 0)
	if err != nil {
		log.Error().Err(err).Msg("ListTeams failed")
	} else {
		fmt.Printf("✅ Found %d teams\n", len(teams.Results))
		if len(teams.Results) > 0 {
			team := teams.Results[0]
			fmt.Printf("   First team: %s (ID: %s)\n", team.Name, team.ID)

			// Test 11: Get team details
			fmt.Println("\n11. Testing GetTeam...")
			teamDetail, err := client.GetTeam(ctx, team.ID)
			if err != nil {
				log.Error().Err(err).Msg("GetTeam failed")
			} else {
				fmt.Printf("✅ Team details: %s\n", teamDetail.Name)
			}
		}
	}

	// Test 12: List files
	if len(workspaces.Results) > 0 {
		workspace := workspaces.Results[0]
		fmt.Println("\n12. Testing ListFiles...")
		files, err := client.ListFiles(ctx, workspace.ID, 10, 0)
		if err != nil {
			log.Error().Err(err).Msg("ListFiles failed")
		} else {
			fmt.Printf("✅ Found %d files\n", len(files.Results))
		}
	}

	fmt.Println("\n🎉 API testing completed!")
	fmt.Println("Check the debug logs above for detailed request/response information.")
}