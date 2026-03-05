package main

import (
	"regexp"
	"strconv"
	"strings"
)

// Priority represents a task's priority level
type Priority int

const (
	PriorityNone    Priority = 0
	PriorityLow     Priority = 1
	PriorityMedium  Priority = 2
	PriorityHigh    Priority = 3
	PriorityHighest Priority = 4
)

// Task represents a parsed Obsidian task
type Task struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	FilePath    string   `json:"filePath"`
	DueDate     string   `json:"dueDate,omitempty"`
	Tags        []string `json:"tags"`
	LineNumber  int      `json:"lineNumber"`
	Priority    Priority `json:"priority"`
}

var (
	taskRegex     = regexp.MustCompile(`^(\s*)- \[([ x])\](.*)$`)
	tagRegex      = regexp.MustCompile(`#[\w-]+`)
	dueDateRegex  = regexp.MustCompile(`(?:📅|🗓️)\s*(\d{4}-\d{2}-\d{2})`)
	priorityRegex = regexp.MustCompile(`[🔺⏫🔼🔽]`)
)

func parsePriority(content string) Priority {
	if strings.Contains(content, "🔺") {
		return PriorityHighest
	}

	if strings.Contains(content, "⏫") {
		return PriorityHigh
	}

	if strings.Contains(content, "🔼") {
		return PriorityMedium
	}

	if strings.Contains(content, "🔽") {
		return PriorityLow
	}

	return PriorityNone
}

// ParseTask parses a markdown task line into a Task struct
func ParseTask(line string, filePath string, lineNumber int) *Task {
	matches := taskRegex.FindStringSubmatch(line)
	if len(matches) < 4 {
		return nil
	}

	status := "incomplete"
	if strings.TrimSpace(matches[2]) == "x" {
		status = "complete"
	}

	content := matches[3]

	// Extract tags
	tags := tagRegex.FindAllString(content, -1)
	if tags == nil {
		tags = []string{}
	}
	// Remove # prefix from tags
	for i, tag := range tags {
		tags[i] = strings.TrimPrefix(tag, "#")
	}

	// Extract due date
	var dueDate string

	dueMatches := dueDateRegex.FindStringSubmatch(content)
	if len(dueMatches) >= 2 {
		dueDate = dueMatches[1]
	}

	// Extract priority
	priority := parsePriority(content)

	// Extract description (remove tags, due date markers, priority emojis)
	description := content
	description = tagRegex.ReplaceAllString(description, "")
	description = dueDateRegex.ReplaceAllString(description, "")
	description = priorityRegex.ReplaceAllString(description, "")
	description = strings.TrimSpace(description)

	id := filePath + ":" + strconv.Itoa(lineNumber)

	return &Task{
		ID:          id,
		Description: description,
		Status:      status,
		FilePath:    filePath,
		LineNumber:  lineNumber,
		Tags:        tags,
		DueDate:     dueDate,
		Priority:    priority,
	}
}
