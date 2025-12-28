package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/worthies/eisenhower/internal/database"
	"github.com/worthies/eisenhower/internal/handlers"
)

const (
	defaultDBPath = "~/.config/eisenhower.db"
)

func main() {
	dbPath := flag.String("db", defaultDBPath, "Path to SQLite database file")
	flag.Parse()

	// Log startup information to stderr (stdout is reserved for MCP protocol)
	log.SetOutput(os.Stderr)
	log.Printf("Eisenhower MCP Server starting...")
	log.Printf("Database path: %s", *dbPath)

	// Initialize database
	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "eisenhower-mcp-server",
		Version: "1.0.0",
	}, nil)

	// Register all tools
	taskHandlers := handlers.NewTaskHandlers(db)
	taskHandlers.RegisterTools(server)

	log.Printf("Server initialized successfully")

	// Start server with stdio transport
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
