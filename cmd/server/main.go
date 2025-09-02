package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/lukasz/nuclino-mcp-server/internal/nuclino"
	"github.com/lukasz/nuclino-mcp-server/internal/server"
)

func main() {
	var (
		debug   = flag.Bool("debug", false, "Enable debug logging")
		version = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *version {
		fmt.Println("nuclino-mcp-server v0.1.0")
		os.Exit(0)
	}

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Debug().Err(err).Msg("No .env file found")
	}

	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Get API key from environment
	apiKey := os.Getenv("NUCLINO_API_KEY")
	if apiKey == "" {
		log.Fatal().Msg("NUCLINO_API_KEY environment variable is required")
	}

	// Create Nuclino client
	nuclinoClient := nuclino.NewClient(apiKey)

	// Create MCP server
	mcpServer := server.NewNuclinoMCPServer(nuclinoClient)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Info().Msg("Received shutdown signal")
		cancel()
	}()

	// Start server
	if err := mcpServer.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
