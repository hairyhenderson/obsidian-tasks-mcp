package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:funlen // comprehensive test cases
func TestParseTask(t *testing.T) {
	tests := []struct {
		want       *Task
		name       string
		line       string
		filePath   string
		lineNumber int
	}{
		{
			name:       "simple incomplete task",
			line:       "- [ ] Buy groceries",
			filePath:   "todo.md",
			lineNumber: 1,
			want: &Task{
				ID:          "todo.md:1",
				Description: "Buy groceries",
				Status:      "incomplete",
				FilePath:    "todo.md",
				LineNumber:  1,
				Tags:        []string{},
				DueDate:     "",
			},
		},
		{
			name:       "simple complete task",
			line:       "- [x] Buy groceries",
			filePath:   "todo.md",
			lineNumber: 2,
			want: &Task{
				ID:          "todo.md:2",
				Description: "Buy groceries",
				Status:      "complete",
				FilePath:    "todo.md",
				LineNumber:  2,
				Tags:        []string{},
				DueDate:     "",
			},
		},
		{
			name:       "task with tags",
			line:       "- [ ] Buy groceries #shopping #urgent",
			filePath:   "todo.md",
			lineNumber: 3,
			want: &Task{
				ID:          "todo.md:3",
				Description: "Buy groceries",
				Status:      "incomplete",
				FilePath:    "todo.md",
				LineNumber:  3,
				Tags:        []string{"shopping", "urgent"},
				DueDate:     "",
			},
		},
		{
			name:       "task with due date calendar emoji",
			line:       "- [ ] Buy groceries 📅 2024-01-15",
			filePath:   "todo.md",
			lineNumber: 4,
			want: &Task{
				ID:          "todo.md:4",
				Description: "Buy groceries",
				Status:      "incomplete",
				FilePath:    "todo.md",
				LineNumber:  4,
				Tags:        []string{},
				DueDate:     "2024-01-15",
			},
		},
		{
			name:       "task with due date calendar2 emoji",
			line:       "- [ ] Buy groceries 🗓️ 2024-01-15",
			filePath:   "todo.md",
			lineNumber: 5,
			want: &Task{
				ID:          "todo.md:5",
				Description: "Buy groceries",
				Status:      "incomplete",
				FilePath:    "todo.md",
				LineNumber:  5,
				Tags:        []string{},
				DueDate:     "2024-01-15",
			},
		},
		{
			name:       "task with tags and due date",
			line:       "- [x] Buy groceries #shopping 📅 2024-01-15",
			filePath:   "todo.md",
			lineNumber: 6,
			want: &Task{
				ID:          "todo.md:6",
				Description: "Buy groceries",
				Status:      "complete",
				FilePath:    "todo.md",
				LineNumber:  6,
				Tags:        []string{"shopping"},
				DueDate:     "2024-01-15",
			},
		},
		{
			name:       "not a task line",
			line:       "This is just regular text",
			filePath:   "todo.md",
			lineNumber: 7,
			want:       nil,
		},
		{
			name:       "task with indentation",
			line:       "  - [ ] Indented task",
			filePath:   "todo.md",
			lineNumber: 8,
			want: &Task{
				ID:          "todo.md:8",
				Description: "Indented task",
				Status:      "incomplete",
				FilePath:    "todo.md",
				LineNumber:  8,
				Tags:        []string{},
				DueDate:     "",
			},
		},
		{
			name:       "task with multiple tags",
			line:       "- [ ] Task #tag1 #tag2 #tag3",
			filePath:   "todo.md",
			lineNumber: 9,
			want: &Task{
				ID:          "todo.md:9",
				Description: "Task",
				Status:      "incomplete",
				FilePath:    "todo.md",
				LineNumber:  9,
				Tags:        []string{"tag1", "tag2", "tag3"},
				DueDate:     "",
			},
		},
		{
			name:       "task with hyphenated tag",
			line:       "- [ ] Task #my-tag",
			filePath:   "todo.md",
			lineNumber: 10,
			want: &Task{
				ID:          "todo.md:10",
				Description: "Task",
				Status:      "incomplete",
				FilePath:    "todo.md",
				LineNumber:  10,
				Tags:        []string{"my-tag"},
				DueDate:     "",
			},
		},
		{
			name:       "task with highest priority emoji",
			line:       "- [ ] Urgent task 🔺",
			filePath:   "todo.md",
			lineNumber: 11,
			want: &Task{
				ID:          "todo.md:11",
				Description: "Urgent task",
				Status:      "incomplete",
				FilePath:    "todo.md",
				LineNumber:  11,
				Tags:        nil,
				DueDate:     "",
				Priority:    PriorityHighest,
			},
		},
		{
			name:       "task with high priority emoji",
			line:       "- [ ] High task ⏫",
			filePath:   "todo.md",
			lineNumber: 12,
			want: &Task{
				ID:          "todo.md:12",
				Description: "High task",
				Status:      "incomplete",
				FilePath:    "todo.md",
				LineNumber:  12,
				Tags:        nil,
				DueDate:     "",
				Priority:    PriorityHigh,
			},
		},
		{
			name:       "task with medium priority emoji",
			line:       "- [ ] Medium task 🔼",
			filePath:   "todo.md",
			lineNumber: 13,
			want: &Task{
				ID:          "todo.md:13",
				Description: "Medium task",
				Status:      "incomplete",
				FilePath:    "todo.md",
				LineNumber:  13,
				Tags:        nil,
				DueDate:     "",
				Priority:    PriorityMedium,
			},
		},
		{
			name:       "task with low priority emoji",
			line:       "- [ ] Low task 🔽",
			filePath:   "todo.md",
			lineNumber: 14,
			want: &Task{
				ID:          "todo.md:14",
				Description: "Low task",
				Status:      "incomplete",
				FilePath:    "todo.md",
				LineNumber:  14,
				Tags:        nil,
				DueDate:     "",
				Priority:    PriorityLow,
			},
		},
		{
			name:       "task with no priority emoji",
			line:       "- [ ] Normal task",
			filePath:   "todo.md",
			lineNumber: 15,
			want: &Task{
				ID:          "todo.md:15",
				Description: "Normal task",
				Status:      "incomplete",
				FilePath:    "todo.md",
				LineNumber:  15,
				Tags:        nil,
				DueDate:     "",
				Priority:    PriorityNone,
			},
		},
		{
			name:       "priority emoji stripped from description",
			line:       "- [ ] Do thing ⏫ #action 📅 2026-03-04",
			filePath:   "todo.md",
			lineNumber: 16,
			want: &Task{
				ID:          "todo.md:16",
				Description: "Do thing",
				Status:      "incomplete",
				FilePath:    "todo.md",
				LineNumber:  16,
				Tags:        []string{"action"},
				DueDate:     "2026-03-04",
				Priority:    PriorityHigh,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTask(tt.line, tt.filePath, tt.lineNumber)

			if tt.want == nil {
				assert.Nil(t, got)

				return
			}

			require.NotNil(t, got)
			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.Description, got.Description)
			assert.Equal(t, tt.want.Status, got.Status)
			assert.Equal(t, tt.want.FilePath, got.FilePath)
			assert.Equal(t, tt.want.LineNumber, got.LineNumber)
			assert.Equal(t, tt.want.Tags, got.Tags)
			assert.Equal(t, tt.want.DueDate, got.DueDate)
			assert.Equal(t, tt.want.Priority, got.Priority)
		})
	}
}
