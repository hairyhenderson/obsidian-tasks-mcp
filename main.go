package main

import (
	"context"
	"flag"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type QueryTasksInput struct {
	Query    string   `json:"query" jsonschema:"Tasks query string with filters (one filter per line). Example: not done\ntag include #shopping"`
	RootDirs []string `json:"rootDirs" jsonschema:"Root directories to scan for markdown files"`
}

type QueryTasksOutput struct {
	Tasks []*Task `json:"tasks"`
}

func queryTasks(ctx context.Context, req *mcp.CallToolRequest, input QueryTasksInput) (
	*mcp.CallToolResult,
	QueryTasksOutput,
	error,
) {
	// Use rootDirs from input, or fall back to command-line args
	roots := input.RootDirs
	if len(roots) == 0 {
		// This shouldn't happen if rootDirs is required, but handle it gracefully
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "rootDirs parameter is required",
				},
			},
		}, QueryTasksOutput{Tasks: []*Task{}}, nil
	}

	// Parse query
	var query *Query
	var err error
	if input.Query != "" {
		query, err = ParseQuery(input.Query)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "failed to parse query: " + err.Error(),
					},
				},
			}, QueryTasksOutput{Tasks: []*Task{}}, nil
		}
	}

	// Scan tasks
	tasks, err := ScanTasksWithQuery(roots, query)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "failed to scan tasks: " + err.Error(),
				},
			},
		}, QueryTasksOutput{Tasks: []*Task{}}, nil
	}

	// Return results (tasks is always a slice, never nil, from ScanTasksWithQuery)
	return nil, QueryTasksOutput{Tasks: tasks}, nil
}

func main() {
	var rootDirs flagList
	flag.Var(&rootDirs, "root", "Root directory to scan for markdown files (can be specified multiple times)")
	flag.Parse()

	if len(rootDirs) == 0 {
		log.Fatal("at least one -root directory must be specified")
	}

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "obsidian-tasks",
		Version: "0.1.0",
	}, nil)

	// Add the query_tasks tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "query_tasks",
		Description: "Query Obsidian tasks from markdown files using Tasks query filters",
	}, queryTasks)

	// Run the server over stdin/stdout
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

// flagList is a custom flag type that allows multiple values
type flagList []string

func (f *flagList) String() string {
	return ""
}

func (f *flagList) Set(value string) error {
	*f = append(*f, value)
	return nil
}
