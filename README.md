# Eisenhower MCP Server

✨✨✨ A full-featured MCP Server for Eisenhower Matrix task management

## Overview

This is a Model Context Protocol (MCP) server that provides comprehensive task management capabilities based on the Eisenhower Matrix (also known as the Urgent-Important Matrix). It helps you prioritize tasks by categorizing them into four quadrants:

- **Urgent & Important** (Do First): Critical tasks requiring immediate attention
- **Not Urgent & Important** (Schedule): Important tasks to plan and schedule
- **Urgent & Not Important** (Delegate): Tasks to delegate if possible
- **Not Urgent & Not Important** (Eliminate): Tasks to eliminate or minimize

## Features

- ✅ Full CRUD operations for tasks
- 📊 Task categorization using Eisenhower Matrix quadrants
- 🔍 Search and filter capabilities
- 📈 Statistics and analytics
- 💾 Persistent storage using SQLite3
- 🔌 Standard MCP protocol support via stdio
- 🛠️ Built with official [github.com/modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk)

👉 **See [FEATURES.md](FEATURES.md) for complete feature list and roadmap**

## Installation

### Prerequisites

- Go 1.21 or higher
- SQLite3

### Build from source

```bash
git clone https://github.com/worthies/eisenhower.git
cd eisenhower
go mod download
go build -o eisenhower
```

Or use the Makefile:

```bash
make build   # Build the server
make install # Build and install to ~/.local/bin
make help    # Show all available commands
```

## Usage

### Running the server

By default, the server stores data in `~/.config/eisenhower.db`:

```bash
./eisenhower-mcp-server
```

To specify a custom database path:

```bash
./eisenhower -db /path/to/your/database.db
```

### VS Code Copilot Integration

Add the following to your VS Code settings (`.vscode/settings.json` or user settings):

```json
{
  "github.copilot.chat.mcp.servers": {
    "eisenhower": {
      "command": "/absolute/path/to/eisenhower-mcp-server",
      "args": ["-db", "~/.config/eisenhower.db"]
    }
  }
}
```

Or in your MCP settings file (`~/.config/Code/User/globalStorage/github.copilot-chat/mcp-servers.json`):

```json
{
  "mcpServers": {
    "eisenhower": {
      "command": "/absolute/path/to/eisenhower",
      "args": ["-db", "~/.config/eisenhower.db"]
    }
  }
}
```

## Available Tools

The server provides the following MCP tools:

### 1. `create_task`

Create a new task in the Eisenhower Matrix.

**Parameters:**

- `title` (required): Task title
- `quadrant` (required): One of: `urgent_important`, `not_urgent_important`, `urgent_not_important`, `not_urgent_not_important`
- `description` (optional): Detailed description
- `priority` (optional): Priority level 1-5 (default: 3)
- `status` (optional): One of: `pending`, `in_progress`, `completed`, `cancelled` (default: `pending`)
- `due_date` (optional): Due date in RFC3339 format (e.g., `2024-12-31T23:59:59Z`)
- `tags` (optional): Comma-separated tags

### 2. `get_task`

Retrieve a specific task by ID.

**Parameters:**

- `id` (required): Task ID

### 3. `list_tasks`

List all tasks with optional filtering.

**Parameters:**

- `quadrant` (optional): Filter by quadrant
- `status` (optional): Filter by status

### 4. `update_task`

Update an existing task.

**Parameters:**

- `id` (required): Task ID
- All other fields from `create_task` are optional

### 5. `delete_task`

Delete a task by ID.

**Parameters:**

- `id` (required): Task ID

### 6. `search_tasks`

Search tasks by title, description, or tags.

**Parameters:**

- `query` (required): Search term

### 7. `move_task`

Move a task to a different quadrant.

**Parameters:**

- `id` (required): Task ID
- `quadrant` (required): Target quadrant

### 8. `complete_task`

Mark a task as completed.

**Parameters:**

- `id` (required): Task ID

### 9. `get_statistics`

Get statistics about all tasks (counts by quadrant and status).

**Parameters:** None

## Example Usage with VS Code Copilot

Once configured, you can interact with the Eisenhower Matrix directly in VS Code Copilot Chat:

```
Create a task "Finish project report" in urgent_important quadrant with high priority
```

```
List all pending tasks in the not_urgent_important quadrant
```

```
Search for tasks related to "meeting"
```

```
Move task 5 to urgent_important quadrant
```

```
Show me statistics about my tasks
```

👉 **See [USAGE.md](USAGE.md) for comprehensive usage guide and examples**

## Database Schema

The server uses a SQLite database with the following schema:

```sql
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    quadrant TEXT NOT NULL CHECK(quadrant IN ('urgent_important', 'not_urgent_important', 'urgent_not_important', 'not_urgent_not_important')),
    priority INTEGER DEFAULT 3 CHECK(priority >= 1 AND priority <= 5),
    status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'completed', 'cancelled')),
    due_date DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    tags TEXT
);
```

## Project Structure

```
eisenhower/
├── main.go                      # Main entry point
├── internal/
│   ├── database/
│   │   └── database.go          # Database layer with SQLite operations
│   └── handlers/
│       └── handlers.go          # MCP tool handlers
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

## Development

### Running tests

```bash
go test ./...
```

### Adding new tools

1. Add the tool definition in `internal/handlers/handlers.go`
2. Implement the handler function
3. Register the tool in the `RegisterTools` method

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## About Eisenhower Matrix

The Eisenhower Matrix, also known as the Urgent-Important Matrix, is a time management framework that helps you prioritize tasks by urgency and importance. It was named after Dwight D. Eisenhower, who reportedly said:

> "What is important is seldom urgent and what is urgent is seldom important."

Learn more: [Eisenhower Matrix on Wikipedia](https://en.wikipedia.org/wiki/Time_management#The_Eisenhower_Method)
