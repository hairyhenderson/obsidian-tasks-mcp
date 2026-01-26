package main

import (
	"testing"
)

func TestParseTask(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		filePath   string
		lineNumber int
		want       *Task
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTask(tt.line, tt.filePath, tt.lineNumber)
			if tt.want == nil {
				if got != nil {
					t.Errorf("ParseTask() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("ParseTask() = nil, want %+v", tt.want)
			}
			if got.ID != tt.want.ID {
				t.Errorf("ParseTask() ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.Description != tt.want.Description {
				t.Errorf("ParseTask() Description = %v, want %v", got.Description, tt.want.Description)
			}
			if got.Status != tt.want.Status {
				t.Errorf("ParseTask() Status = %v, want %v", got.Status, tt.want.Status)
			}
			if got.FilePath != tt.want.FilePath {
				t.Errorf("ParseTask() FilePath = %v, want %v", got.FilePath, tt.want.FilePath)
			}
			if got.LineNumber != tt.want.LineNumber {
				t.Errorf("ParseTask() LineNumber = %v, want %v", got.LineNumber, tt.want.LineNumber)
			}
			if len(got.Tags) != len(tt.want.Tags) {
				t.Errorf("ParseTask() Tags length = %v, want %v", len(got.Tags), len(tt.want.Tags))
			} else {
				for i, tag := range got.Tags {
					if tag != tt.want.Tags[i] {
						t.Errorf("ParseTask() Tags[%d] = %v, want %v", i, tag, tt.want.Tags[i])
					}
				}
			}
			if got.DueDate != tt.want.DueDate {
				t.Errorf("ParseTask() DueDate = %v, want %v", got.DueDate, tt.want.DueDate)
			}
		})
	}
}
