package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanTasks(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	
	// Create test markdown files
	file1 := filepath.Join(tmpDir, "todo.md")
	err := os.WriteFile(file1, []byte(`# Tasks

- [ ] Buy groceries #shopping
- [x] Done task
- [ ] Task with due date 📅 2024-01-15
`), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	
	file2 := filepath.Join(tmpDir, "notes", "project.md")
	err = os.MkdirAll(filepath.Dir(file2), 0755)
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	err = os.WriteFile(file2, []byte(`# Project Tasks

- [ ] Task in notes #project
`), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	
	// Scan tasks
	tasks, err := ScanTasks([]string{tmpDir})
	if err != nil {
		t.Fatalf("ScanTasks() error = %v", err)
	}
	
	// Should find 4 tasks
	if len(tasks) != 4 {
		t.Errorf("ScanTasks() found %d tasks, want 4", len(tasks))
	}
	
	// Verify tasks are sorted by file path then line number
	// File paths should be relative to the root directory
	expectedOrder := []struct {
		filePath   string
		lineNumber int
		status     string
	}{
		{"notes/project.md", 3, "incomplete"},
		{"todo.md", 3, "incomplete"},
		{"todo.md", 4, "complete"},
		{"todo.md", 5, "incomplete"},
	}
	
	for i, exp := range expectedOrder {
		if i >= len(tasks) {
			t.Fatalf("not enough tasks, expected at least %d", i+1)
		}
		if tasks[i].FilePath != exp.filePath {
			t.Errorf("task %d: FilePath = %v, want %v", i, tasks[i].FilePath, exp.filePath)
		}
		if tasks[i].LineNumber != exp.lineNumber {
			t.Errorf("task %d: LineNumber = %v, want %v", i, tasks[i].LineNumber, exp.lineNumber)
		}
		if tasks[i].Status != exp.status {
			t.Errorf("task %d: Status = %v, want %v", i, tasks[i].Status, exp.status)
		}
	}
}

func TestScanTasksWithQuery(t *testing.T) {
	tmpDir := t.TempDir()
	
	file1 := filepath.Join(tmpDir, "todo.md")
	err := os.WriteFile(file1, []byte(`# Tasks

- [ ] Buy groceries #shopping
- [x] Done task #shopping
- [ ] Other task
`), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	
	// Query for incomplete tasks with shopping tag
	query, err := ParseQuery("not done\ntag include #shopping")
	if err != nil {
		t.Fatalf("ParseQuery() error = %v", err)
	}
	
	tasks, err := ScanTasksWithQuery([]string{tmpDir}, query)
	if err != nil {
		t.Fatalf("ScanTasksWithQuery() error = %v", err)
	}
	
	// Should find 1 task
	if len(tasks) != 1 {
		t.Errorf("ScanTasksWithQuery() found %d tasks, want 1", len(tasks))
	}
	
	if tasks[0].Description != "Buy groceries" {
		t.Errorf("task description = %v, want 'Buy groceries'", tasks[0].Description)
	}
}
