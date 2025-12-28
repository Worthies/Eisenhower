# Eisenhower MCP Server - Usage Examples

## Quick Start

### 1. Build and Install

```bash
# Build the server
make build

# Or install to ~/.local/bin
make install
```

### 2. Configure VS Code

Add this to your VS Code settings (`.vscode/settings.json` or User Settings):

```json
{
  "github.copilot.chat.mcp.servers": {
    "eisenhower": {
      "command": "/absolute/path/to/eisenhower",
      "args": ["-db", "~/.config/eisenhower.db"]
    }
  }
}
```

If you used `make install`:

```json
{
  "github.copilot.chat.mcp.servers": {
    "eisenhower": {
      "command": "/home/yourusername/.local/bin/eisenhower",
      "args": ["-db", "~/.config/eisenhower.db"]
    }
  }
}
```

### 3. Restart VS Code

After adding the configuration, restart VS Code to load the MCP server.

## Usage in VS Code Copilot Chat

Once configured, you can interact with the Eisenhower Matrix directly
in Copilot Chat:

### Creating Tasks

**Example 1: Create an urgent and important task**

```
Create a task "Fix critical production bug" in urgent_important quadrant
with priority 5
```

**Example 2: Create a task with due date**

```
Create a task "Review Q1 reports" in not_urgent_important quadrant with
due date 2025-03-31T17:00:00Z
```

**Example 3: Create a task with tags**

```
Create a task "Weekly team meeting" in urgent_not_important quadrant with tags "meeting,recurring,team"
```

### Listing Tasks

**Example 1: List all tasks**

```
List all my tasks
```

**Example 2: List tasks in specific quadrant**

```
Show me all tasks in the urgent_important quadrant
```

**Example 3: List pending tasks**

```
List all pending tasks
```

**Example 4: List completed tasks in specific quadrant**

```
Show completed tasks in not_urgent_important quadrant
```

### Updating Tasks

**Example 1: Update task title**

```
Update task 3 with new title "Refactor authentication module"
```

**Example 2: Change task priority**

```
Set priority of task 5 to 4
```

**Example 3: Update multiple fields**

```
Update task 2: set status to in_progress and priority to 5
```

### Managing Tasks

**Example 1: Move task to different quadrant**

```
Move task 4 to urgent_important quadrant
```

**Example 2: Complete a task**

```
Mark task 7 as completed
```

**Example 3: Delete a task**

```
Delete task 9
```

### Searching Tasks

**Example 1: Search by keyword**

```
Search for tasks containing "report"
```

**Example 2: Search in tags**

```
Find all tasks tagged with "meeting"
```

### Getting Statistics

```
Show me statistics about my tasks
```

This will return:

- Total number of tasks
- Tasks count by quadrant
- Tasks count by status

## Understanding the Quadrants

### 1. Urgent & Important (Do First)

- Critical deadlines
- Emergency situations
- Crisis management

### 2. Not Urgent & Important (Schedule)

- Long-term planning
- Personal development
- Relationship building
- Strategic work

### 3. Urgent & Not Important (Delegate)

- Interruptions
- Some emails/calls
- Some meetings
- Other people's priorities

### 4. Not Urgent & Not Important (Eliminate)

- Time wasters
- Busy work
- Trivial tasks
- Some entertainment

## Task Properties

Each task has the following properties:

- **ID**: Auto-generated unique identifier
- **Title**: Task name (required)
- **Description**: Detailed description (optional)
- **Quadrant**: One of the four quadrants (required)
- **Priority**: 1-5, where 5 is highest (default: 3)
- **Status**: pending, in_progress, completed, or cancelled (default: pending)
- **Due Date**: RFC3339 format timestamp (optional)
- **Tags**: Comma-separated tags for categorization (optional)
- **Created At**: Auto-generated timestamp
- **Updated At**: Auto-updated timestamp

## Tips for Best Results

1. **Be specific with quadrants**: Always mention the quadrant name when creating tasks
2. **Use priority wisely**: Reserve priority 5 for truly critical tasks
3. **Tag consistently**: Use consistent tag names for better organization
4. **Review regularly**: Periodically check your task statistics and adjust priorities

## Troubleshooting

### Server not responding

1. Check if the server path in settings is correct
2. Verify the database path is writable
3. Check VS Code's Output panel for errors

### Database issues

- Default database location: `~/.config/eisenhower.db`
- Database is created automatically on first run
- To reset: delete the database file and restart

### Server logs

Server logs are written to stderr. To see them:

```bash
./eisenhower -db /tmp/test.db 2>&1 | tee server.log
```
