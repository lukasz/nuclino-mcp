package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/lukasz/nuclino-mcp-server/internal/nuclino"
)

func main() {
	// Setup logging
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	apiKey := os.Getenv("NUCLINO_API_KEY")
	if apiKey == "" {
		log.Fatal().Msg("NUCLINO_API_KEY required")
	}

	client := nuclino.NewClient(apiKey)
	ctx := context.Background()

	fmt.Println("1. Testing ListWorkspaces...")
	workspaces, err := client.ListWorkspaces(ctx, 5, 0)
	if err != nil {
		log.Error().Err(err).Msg("Failed")
		return
	}

	fmt.Printf("Found %d workspaces\n", len(workspaces.Results))
	if len(workspaces.Results) == 0 {
		fmt.Println("No workspaces found")
		return
	}

	workspace := workspaces.Results[0]
	fmt.Printf("First workspace: ID=%s, Name='%s'\n", workspace.ID, workspace.Name)

	fmt.Println("\n2. Testing GetWorkspace...")
	ws, err := client.GetWorkspace(ctx, workspace.ID)
	if err != nil {
		log.Error().Err(err).Msg("GetWorkspace failed")
		return
	}

	fmt.Printf("Workspace details: ID='%s', Name='%s', TeamID='%s'\n", 
		ws.ID, ws.Name, ws.TeamID)

	// Check if fields are empty
	if ws.ID == "" && ws.Name == "" && ws.TeamID == "" {
		fmt.Println("⚠️  WARNING: All fields are empty! Response parsing issue.")
	}

	fmt.Println("\n3. Testing ListItems...")
	items, err := client.ListItems(ctx, workspace.ID, 5, 0)
	if err != nil {
		log.Error().Err(err).Msg("ListItems failed")
		return
	}

	fmt.Printf("Found %d items\n", len(items.Results))

	fmt.Println("\n4. Testing ListTeams...")
	teams, err := client.ListTeams(ctx, 5, 0) 
	if err != nil {
		log.Error().Err(err).Msg("ListTeams failed")
		return
	}

	fmt.Printf("Found %d teams\n", len(teams.Results))
	if len(teams.Results) > 0 {
		team := teams.Results[0]
		fmt.Printf("First team: ID=%s, Name='%s'\n", team.ID, team.Name)

		fmt.Println("\n5. Testing GetTeam...")
		teamDetail, err := client.GetTeam(ctx, team.ID)
		if err != nil {
			log.Error().Err(err).Msg("GetTeam failed")
			return
		}
		
		fmt.Printf("Team details: ID='%s', Name='%s'\n", teamDetail.ID, teamDetail.Name)

		if teamDetail.ID == "" && teamDetail.Name == "" {
			fmt.Println("⚠️  WARNING: Team fields are empty! Response parsing issue.")
		}
	}

	fmt.Println("\n✅ Quick test completed!")
}