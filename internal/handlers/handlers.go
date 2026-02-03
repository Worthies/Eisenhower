package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/worthies/eisenhower/internal/database"
)

// Define MCP tool argument types at package level for proper schema generation
type CreateTaskArgs struct {
	Title       string  `json:"title" jsonschema:"The title of the task"`
	Description string  `json:"description" jsonschema:"Detailed description of the task"`
	Quadrant    string  `json:"quadrant" jsonschema:"The Eisenhower Matrix quadrant: urgent_important, not_urgent_important, urgent_not_important, or not_urgent_not_important"`
	Priority    int     `json:"priority" jsonschema:"Priority level from 1 to 5, where 5 is highest priority. Default is 3"`
	Status      string  `json:"status" jsonschema:"Task status: pending, in_progress, completed, or cancelled. Default is pending"`
	DueDate     *string `json:"due_date" jsonschema:"Due date in RFC3339 format, e.g. 2024-12-31T23:59:59Z"`
	StartedAt   *string `json:"started_at" jsonschema:"Start timestamp in RFC3339 format, e.g. 2024-12-31T23:59:59Z"`
	Tags        string  `json:"tags" jsonschema:"Comma-separated tags for categorizing the task"`
	Progress    *int    `json:"progress" jsonschema:"Progress percentage (0-100)"`
	Summary     string  `json:"summary" jsonschema:"Initial progress summary information"`
	Project     *string `json:"project" jsonschema:"Project this task belongs to"`
}

type UpdateTaskArgs struct {
	ID          int64   `json:"id" jsonschema:"The task ID"`
	Title       *string `json:"title" jsonschema:"The title of the task"`
	Description *string `json:"description" jsonschema:"Detailed description of the task"`
	Quadrant    *string `json:"quadrant" jsonschema:"The Eisenhower Matrix quadrant"`
	Priority    *int    `json:"priority" jsonschema:"Priority level from 1 to 5"`
	Status      *string `json:"status" jsonschema:"Task status"`
	DueDate     *string `json:"due_date" jsonschema:"Due date in RFC3339 format"`
	StartedAt   *string `json:"started_at" jsonschema:"Start timestamp in RFC3339 format"`
	Tags        *string `json:"tags" jsonschema:"Comma-separated tags"`
	Progress    *int    `json:"progress" jsonschema:"Progress percentage (0-100)"`
	Summary     *string `json:"summary" jsonschema:"Progress information to insert to existing summary"`
	Project     *string `json:"project" jsonschema:"Project this task belongs to"`
}

type GetProjectsArgs struct {
	Status *string `json:"status" jsonschema:"Filter by status: pending, in_progress, completed, cancelled. If not provided, returns projects from all tasks."`
}

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
	// Create Task - manually build schema to ensure all fields are included
	createTaskSchema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"title":       {Type: "string", Description: "The title of the task"},
			"description": {Type: "string", Description: "Detailed description of the task"},
			"quadrant":    {Type: "string", Description: "The Eisenhower Matrix quadrant"},
			"priority":    {Type: "integer", Description: "Priority level from 1 to 5"},
			"status":      {Type: "string", Description: "Task status"},
			"due_date":    {Type: "string", Description: "Due date in RFC3339 format"},
			"started_at":  {Type: "string", Description: "Start timestamp in RFC3339 format"},
			"tags":        {Type: "string", Description: "Comma-separated tags"},
			"progress":    {Type: "integer", Description: "Progress percentage"},
			"summary":     {Type: "string", Description: "Summary"},
			"project":     {Type: "string", Description: "Project this task belongs to"},
		},
		Required: []string{"title", "quadrant"},
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_task",
		Description: "Create a new task in the Eisenhower Matrix. Quadrants: urgent_important (Do First), not_urgent_important (Schedule), urgent_not_important (Delegate), not_urgent_not_important (Eliminate).",
		InputSchema: createTaskSchema,
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateTaskArgs) (*mcp.CallToolResult, any, error) {
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
		// Set project with default value
		if args.Project != nil && *args.Project != "" {
			task.Project = args.Project
		} else {
			defaultProject := "Routine"
			task.Project = &defaultProject
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

		if args.StartedAt != nil && *args.StartedAt != "" {
			startedAt, err := time.Parse(time.RFC3339, *args.StartedAt)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid started_at format: %w", err)
			}
			task.StartedAt = &startedAt
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
		Project  *string `json:"project,omitempty" jsonschema:"Filter by project name"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_tasks",
		Description: "List all tasks with optional filtering by quadrant, status, and/or project.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args listTasksArgs) (*mcp.CallToolResult, any, error) {
		var quadrant *database.Quadrant
		if args.Quadrant != nil {
			q := database.Quadrant(*args.Quadrant)
			quadrant = &q
		}

		tasks, err := h.db.ListTasks(quadrant, args.Status, args.Project)
		if err != nil {
			return nil, nil, err
		}

		return formatResponse(tasks), nil, nil
	})

	// Update Task - manually build schema to ensure all fields are included
	updateTaskSchema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"id":          {Type: "integer", Description: "The task ID"},
			"title":       {Type: "string", Description: "The title of the task"},
			"description": {Type: "string", Description: "Detailed description of the task"},
			"quadrant":    {Type: "string", Description: "The Eisenhower Matrix quadrant"},
			"priority":    {Type: "integer", Description: "Priority level from 1 to 5"},
			"status":      {Type: "string", Description: "Task status"},
			"due_date":    {Type: "string", Description: "Due date in RFC3339 format"},
			"started_at":  {Type: "string", Description: "Start timestamp in RFC3339 format"},
			"tags":        {Type: "string", Description: "Comma-separated tags"},
			"progress":    {Type: "integer", Description: "Progress percentage"},
			"summary":     {Type: "string", Description: "Summary"},
			"project":     {Type: "string", Description: "Project this task belongs to"},
		},
		Required: []string{"id"},
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_task",
		Description: "Update an existing task. All fields except ID are optional - only provide the fields you want to update.",
		InputSchema: updateTaskSchema,
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateTaskArgs) (*mcp.CallToolResult, any, error) {
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
			// Set finish timestamp when task is completed or cancelled
			if *args.Status == "completed" || *args.Status == "cancelled" {
				now := time.Now()
				task.FinishedAt = &now
			}
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
		if args.StartedAt != nil {
			startedAt, err := time.Parse(time.RFC3339, *args.StartedAt)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid started_at format: %w", err)
			}
			task.StartedAt = &startedAt
		}
		if args.Project != nil && *args.Project != "" {
			task.Project = args.Project
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
		now := time.Now()
		task.FinishedAt = &now

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

	// Get Projects
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_projects",
		Description: "Get unique project names from all tasks. Optionally filter to show only projects with active (pending/in_progress) tasks.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"include_completed": map[string]interface{}{
					"type":        "boolean",
					"description": "If true, include projects where all tasks are completed or cancelled. If false, only show projects with at least one pending or in_progress task. Default is true (show all projects).",
				},
			},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetProjectsArgs) (*mcp.CallToolResult, any, error) {
		includeCompleted := true // default

		// Parse the include_completed parameter from raw JSON
		if len(req.Params.Arguments) > 0 {
			var argMap map[string]interface{}
			if err := json.Unmarshal(req.Params.Arguments, &argMap); err == nil {
				if val, exists := argMap["include_completed"]; exists {
					if boolVal, ok := val.(bool); ok {
						includeCompleted = boolVal
					}
				}
			}
		}

		projects, err := h.db.GetDistinctProjects(includeCompleted)
		if err != nil {
			return nil, nil, err
		}

		// If no projects found, return empty array
		if projects == nil {
			projects = []string{}
		}

		response := map[string]interface{}{
			"projects": projects,
			"count":    len(projects),
		}

		return formatResponse(response), nil, nil
	})
}

// Helper function to format responses as JSON
func formatResponse(data interface{}) *mcp.CallToolResult {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		// Fallback to fmt.Sprintf if JSON marshaling fails
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("%+v", data),
				},
			},
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}
}
