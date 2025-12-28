package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/worthies/eisenhower/internal/database"
)

// TaskHandlers contains handlers for all task-related MCP tools
type TaskHandlers struct {
	db *database.DB
}

// NewTaskHandlers creates a new TaskHandlers instance
func NewTaskHandlers(db *database.DB) *TaskHandlers {
	return &TaskHandlers{db: db}
}

// RegisterTools registers all task management tools with the MCP server
func (h *TaskHandlers) RegisterTools(s *mcp.Server) {
	// Create Task
	type createTaskArgs struct {
		Title       string  `json:"title" jsonschema:"The title of the task"`
		Description string  `json:"description,omitempty" jsonschema:"Detailed description of the task"`
		Quadrant    string  `json:"quadrant" jsonschema:"The Eisenhower Matrix quadrant: urgent_important, not_urgent_important, urgent_not_important, or not_urgent_not_important"`
		Priority    int     `json:"priority,omitempty" jsonschema:"Priority level from 1 to 5, where 5 is highest priority. Default is 3"`
		Status      string  `json:"status,omitempty" jsonschema:"Task status: pending, in_progress, completed, or cancelled. Default is pending"`
		DueDate     *string `json:"due_date,omitempty" jsonschema:"Due date in RFC3339 format, e.g. 2024-12-31T23:59:59Z"`
		Tags        string  `json:"tags,omitempty" jsonschema:"Comma-separated tags for categorizing the task"`
		Progress    *int    `json:"progress,omitempty" jsonschema:"Progress percentage (0-100)"`
		Summary     string  `json:"summary,omitempty" jsonschema:"Initial progress summary information"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_task",
		Description: "Create a new task in the Eisenhower Matrix. Quadrants: urgent_important (Do First), not_urgent_important (Schedule), urgent_not_important (Delegate), not_urgent_not_important (Eliminate).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args createTaskArgs) (*mcp.CallToolResult, any, error) {
		task := &database.Task{
			Title:       args.Title,
			Description: args.Description,
			Quadrant:    database.Quadrant(args.Quadrant),
			Priority:    args.Priority,
			Status:      args.Status,
			Tags:        args.Tags,
			Summary:     args.Summary,
		}

		if task.Priority == 0 {
			task.Priority = 3
		}
		if task.Status == "" {
			task.Status = "pending"
		}
		if args.Progress != nil {
			task.Progress = *args.Progress
		}

		if args.DueDate != nil && *args.DueDate != "" {
			dueDate, err := time.Parse(time.RFC3339, *args.DueDate)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid due_date format: %w", err)
			}
			task.DueDate = &dueDate
		}

		if err := h.db.CreateTask(task); err != nil {
			return nil, nil, err
		}

		return formatResponse(task), nil, nil
	})

	// Get Task
	type getTaskArgs struct {
		ID int64 `json:"id" jsonschema:"The task ID"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_task",
		Description: "Retrieve a specific task by its ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args getTaskArgs) (*mcp.CallToolResult, any, error) {
		task, err := h.db.GetTask(args.ID)
		if err != nil {
			return nil, nil, err
		}
		return formatResponse(task), nil, nil
	})

	// List Tasks
	type listTasksArgs struct {
		Quadrant *string `json:"quadrant,omitempty" jsonschema:"Filter by quadrant: urgent_important, not_urgent_important, urgent_not_important, or not_urgent_not_important"`
		Status   *string `json:"status,omitempty" jsonschema:"Filter by status: pending, in_progress, completed, or cancelled"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_tasks",
		Description: "List all tasks with optional filtering by quadrant and/or status.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args listTasksArgs) (*mcp.CallToolResult, any, error) {
		var quadrant *database.Quadrant
		if args.Quadrant != nil {
			q := database.Quadrant(*args.Quadrant)
			quadrant = &q
		}

		tasks, err := h.db.ListTasks(quadrant, args.Status)
		if err != nil {
			return nil, nil, err
		}

		return formatResponse(tasks), nil, nil
	})

	// Update Task
	type updateTaskArgs struct {
		ID          int64   `json:"id" jsonschema:"The task ID"`
		Title       *string `json:"title,omitempty" jsonschema:"The title of the task"`
		Description *string `json:"description,omitempty" jsonschema:"Detailed description of the task"`
		Quadrant    *string `json:"quadrant,omitempty" jsonschema:"The Eisenhower Matrix quadrant"`
		Priority    *int    `json:"priority,omitempty" jsonschema:"Priority level from 1 to 5"`
		Status      *string `json:"status,omitempty" jsonschema:"Task status"`
		DueDate     *string `json:"due_date,omitempty" jsonschema:"Due date in RFC3339 format"`
		Tags        *string `json:"tags,omitempty" jsonschema:"Comma-separated tags"`
		Progress    *int    `json:"progress,omitempty" jsonschema:"Progress percentage (0-100)"`
		Summary     *string `json:"summary,omitempty" jsonschema:"Progress information to insert to existing summary"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_task",
		Description: "Update an existing task. All fields except ID are optional - only provide the fields you want to update.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args updateTaskArgs) (*mcp.CallToolResult, any, error) {
		task, err := h.db.GetTask(args.ID)
		if err != nil {
			return nil, nil, err
		}

		// Update only provided fields
		if args.Title != nil {
			task.Title = *args.Title
		}
		if args.Description != nil {
			task.Description = *args.Description
		}
		if args.Quadrant != nil {
			task.Quadrant = database.Quadrant(*args.Quadrant)
		}
		if args.Priority != nil {
			task.Priority = *args.Priority
		}
		if args.Status != nil {
			task.Status = *args.Status
		}
		if args.Tags != nil {
			task.Tags = *args.Tags
		}
		if args.Progress != nil {
			task.Progress = *args.Progress
		}
		if args.Summary != nil && *args.Summary != "" {
			// Prepend new summary with timestamp to existing summary
			timestamp := time.Now().Format(time.RFC3339)
			newSummaryEntry := fmt.Sprintf("[%s] %s", timestamp, *args.Summary)
			if task.Summary != "" {
				task.Summary = newSummaryEntry + "\n" + task.Summary
			} else {
				task.Summary = newSummaryEntry
			}
		}
		if args.DueDate != nil {
			dueDate, err := time.Parse(time.RFC3339, *args.DueDate)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid due_date format: %w", err)
			}
			task.DueDate = &dueDate
		}

		if err := h.db.UpdateTask(task); err != nil {
			return nil, nil, err
		}

		return formatResponse(task), nil, nil
	})

	// Delete Task
	type deleteTaskArgs struct {
		ID int64 `json:"id" jsonschema:"The task ID"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_task",
		Description: "Delete a task by its ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args deleteTaskArgs) (*mcp.CallToolResult, any, error) {
		if err := h.db.DeleteTask(args.ID); err != nil {
			return nil, nil, err
		}

		return formatResponse(map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Task %d deleted successfully", args.ID),
		}), nil, nil
	})

	// Search Tasks
	type searchTasksArgs struct {
		Query string `json:"query" jsonschema:"Search term to find in task title, description, or tags"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_tasks",
		Description: "Search tasks by title, description, or tags.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args searchTasksArgs) (*mcp.CallToolResult, any, error) {
		tasks, err := h.db.SearchTasks(args.Query)
		if err != nil {
			return nil, nil, err
		}

		return formatResponse(tasks), nil, nil
	})

	// Move Task
	type moveTaskArgs struct {
		ID       int64  `json:"id" jsonschema:"The task ID"`
		Quadrant string `json:"quadrant" jsonschema:"Target quadrant to move the task to"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "move_task",
		Description: "Move a task to a different quadrant in the Eisenhower Matrix.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args moveTaskArgs) (*mcp.CallToolResult, any, error) {
		task, err := h.db.GetTask(args.ID)
		if err != nil {
			return nil, nil, err
		}

		task.Quadrant = database.Quadrant(args.Quadrant)

		if err := h.db.UpdateTask(task); err != nil {
			return nil, nil, err
		}

		return formatResponse(task), nil, nil
	})

	// Complete Task
	type completeTaskArgs struct {
		ID int64 `json:"id" jsonschema:"The task ID"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "complete_task",
		Description: "Mark a task as completed.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args completeTaskArgs) (*mcp.CallToolResult, any, error) {
		task, err := h.db.GetTask(args.ID)
		if err != nil {
			return nil, nil, err
		}

		task.Status = "completed"

		if err := h.db.UpdateTask(task); err != nil {
			return nil, nil, err
		}

		return formatResponse(task), nil, nil
	})

	// Get Statistics
	type getStatisticsArgs struct{}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_statistics",
		Description: "Get statistics about all tasks, including counts by quadrant and status.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args getStatisticsArgs) (*mcp.CallToolResult, any, error) {
		stats, err := h.db.GetStatistics()
		if err != nil {
			return nil, nil, err
		}

		return formatResponse(stats), nil, nil
	})
}

// Helper function to format responses
func formatResponse(data interface{}) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("%+v", data),
			},
		},
	}
}
