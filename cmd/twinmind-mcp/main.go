package main

import (
	"log"

	"github.com/jlentink/twinmind-mcp/internal/api"
	"github.com/jlentink/twinmind-mcp/internal/config"
	"github.com/jlentink/twinmind-mcp/internal/mcp"
)

var Version = "dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	client := api.NewClient(cfg)
	srv := mcp.New(client)

	if err := srv.Serve(); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}
