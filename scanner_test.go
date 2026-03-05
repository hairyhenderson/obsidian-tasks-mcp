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
`), 0o600)
	require.NoError(t, err)

	file2 := filepath.Join(tmpDir, "notes", "project.md")
	err = os.MkdirAll(filepath.Dir(file2), 0o755)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte(`# Project Tasks

- [ ] Task in notes #project
`), 0o600)
	require.NoError(t, err)

	// Scan tasks
	tasks, err := ScanTasks([]string{tmpDir})
	require.NoError(t, err)

	// Should find 4 tasks
	assert.Len(t, tasks, 4)

	// Verify tasks are sorted by file path then line number
	// File paths should be relative to the root directory
	//nolint:govet // test struct field alignment not critical
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
`), 0o600)
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

func TestScanTasksWithQuerySortByPriority(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "todo.md")
	err := os.WriteFile(file1, []byte(`# Tasks

- [ ] Low priority task 🔽
- [ ] High priority task ⏫
- [ ] Normal task
- [ ] Highest priority task 🔺
- [ ] Medium priority task 🔼
`), 0o600)
	require.NoError(t, err)

	query := &Query{
		Filters: []Filter{&StatusFilter{Done: false}},
		SortBy:  []SortKey{{Field: SortByPriority, Reverse: true}},
	}

	tasks, err := ScanTasksWithQuery([]string{tmpDir}, query)
	require.NoError(t, err)
	require.Len(t, tasks, 5)

	assert.Equal(t, PriorityHighest, tasks[0].Priority)
	assert.Equal(t, PriorityHigh, tasks[1].Priority)
	assert.Equal(t, PriorityMedium, tasks[2].Priority)
	assert.Equal(t, PriorityLow, tasks[3].Priority)
	assert.Equal(t, PriorityNone, tasks[4].Priority)
}

func TestScanTasksWithQuerySortByDue(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "todo.md")
	err := os.WriteFile(file1, []byte(`# Tasks

- [ ] Task C 📅 2026-03-01
- [ ] Task A 📅 2026-01-01
- [ ] Task no due
- [ ] Task B 📅 2026-02-01
`), 0o600)
	require.NoError(t, err)

	// sort by due reverse = newest first
	query := &Query{
		Filters: []Filter{&StatusFilter{Done: false}},
		SortBy:  []SortKey{{Field: SortByDue, Reverse: true}},
	}

	tasks, err := ScanTasksWithQuery([]string{tmpDir}, query)
	require.NoError(t, err)
	require.Len(t, tasks, 4)

	// Tasks with no due date sort last
	assert.Equal(t, "2026-03-01", tasks[0].DueDate)
	assert.Equal(t, "2026-02-01", tasks[1].DueDate)
	assert.Equal(t, "2026-01-01", tasks[2].DueDate)
	assert.Equal(t, "", tasks[3].DueDate)
}

func TestScanTasksWithQueryLimit(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "todo.md")
	err := os.WriteFile(file1, []byte(`# Tasks

- [ ] Task one
- [ ] Task two
- [ ] Task three
- [ ] Task four
- [ ] Task five
`), 0o600)
	require.NoError(t, err)

	query := &Query{
		Filters: []Filter{&StatusFilter{Done: false}},
		Limit:   3,
	}

	tasks, err := ScanTasksWithQuery([]string{tmpDir}, query)
	require.NoError(t, err)

	assert.Len(t, tasks, 3)
}
