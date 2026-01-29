package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
`), 0o644)
	require.NoError(t, err)
	
	file2 := filepath.Join(tmpDir, "notes", "project.md")
	err = os.MkdirAll(filepath.Dir(file2), 0o755)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte(`# Project Tasks

- [ ] Task in notes #project
`), 0o644)
	require.NoError(t, err)
	
	// Scan tasks
	tasks, err := ScanTasks([]string{tmpDir})
	require.NoError(t, err)
	
	// Should find 4 tasks
	assert.Len(t, tasks, 4)
	
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
		require.Less(t, i, len(tasks), "not enough tasks")
		assert.Equal(t, exp.filePath, tasks[i].FilePath)
		assert.Equal(t, exp.lineNumber, tasks[i].LineNumber)
		assert.Equal(t, exp.status, tasks[i].Status)
	}
}

func TestScanTasksWithQuery(t *testing.T) {
	tmpDir := t.TempDir()
	
	file1 := filepath.Join(tmpDir, "todo.md")
	err := os.WriteFile(file1, []byte(`# Tasks

- [ ] Buy groceries #shopping
- [x] Done task #shopping
- [ ] Other task
`), 0o644)
	require.NoError(t, err)
	
	// Query for incomplete tasks with shopping tag
	query, err := ParseQuery("not done\ntag include #shopping")
	require.NoError(t, err)
	
	tasks, err := ScanTasksWithQuery([]string{tmpDir}, query)
	require.NoError(t, err)
	
	// Should find 1 task
	require.Len(t, tasks, 1)
	assert.Equal(t, "Buy groceries", tasks[0].Description)
}
