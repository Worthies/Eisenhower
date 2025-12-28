package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/worthies/eisenhower/internal/database"
	"github.com/worthies/eisenhower/internal/handlers"
)

const (
	defaultDBPath = "~/.config/eisenhower.db"
)

// autoFlushWriter wraps os.File and syncs after each write
type autoFlushWriter struct {
	file *os.File
}

func (w *autoFlushWriter) Write(p []byte) (n int, err error) {
	n, err = w.file.Write(p)
	if err == nil {
		w.file.Sync()
	}
	return
}

func initLogger() error {
	// Expand ~ to home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	logDir := filepath.Join(home, "logs")
	logPath := filepath.Join(logDir, "eisenhower.log")

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// Truncate the log file on startup
	f, err := os.Create(logPath)
	if err != nil {
		return err
	}

	// Set log output to auto-flushing writer
	log.SetOutput(io.Writer(&autoFlushWriter{file: f}))
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	return nil
}

func main() {
	// Initialize logger before any logging
	if err := initLogger(); err != nil {
		// Fallback to stderr if logger initialization fails
		log.SetOutput(os.Stderr)
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	dbPath := flag.String("db", defaultDBPath, "Path to SQLite database file")
	flag.Parse()

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
