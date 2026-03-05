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

	tasks, total, err := ScanTasksWithQuery([]string{tmpDir}, query)
	require.NoError(t, err)

	// Should find 1 task
	require.Len(t, tasks, 1)
	assert.Equal(t, "Buy groceries", tasks[0].Description)
	assert.Equal(t, 1, total)
}

//nolint:funlen // comprehensive pagination test cases
func TestScanTasksWithQuery_OffsetAndLimit(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "todo.md")
	err := os.WriteFile(file1, []byte(`# Tasks

- [ ] Task 1
- [ ] Task 2
- [ ] Task 3
- [ ] Task 4
- [ ] Task 5
`), 0o600)
	require.NoError(t, err)

	t.Run("offset and limit", func(t *testing.T) {
		query := &Query{Offset: 2, Limit: 2}
		tasks, total, err := ScanTasksWithQuery([]string{tmpDir}, query)
		require.NoError(t, err)
		assert.Equal(t, 5, total)
		require.Len(t, tasks, 2)
		assert.Equal(t, "Task 3", tasks[0].Description)
		assert.Equal(t, "Task 4", tasks[1].Description)
	})

	t.Run("offset beyond total", func(t *testing.T) {
		query := &Query{Offset: 10}
		tasks, total, err := ScanTasksWithQuery([]string{tmpDir}, query)
		require.NoError(t, err)
		assert.Equal(t, 5, total)
		assert.Empty(t, tasks)
	})

	t.Run("no offset returns all with total", func(t *testing.T) {
		tasks, total, err := ScanTasksWithQuery([]string{tmpDir}, nil)
		require.NoError(t, err)
		assert.Equal(t, 5, total)
		assert.Len(t, tasks, 5)
	})

	t.Run("limit only", func(t *testing.T) {
		query := &Query{Limit: 3}
		tasks, total, err := ScanTasksWithQuery([]string{tmpDir}, query)
		require.NoError(t, err)
		assert.Equal(t, 5, total)
		require.Len(t, tasks, 3)
		assert.Equal(t, "Task 1", tasks[0].Description)
	})

	t.Run("offset equals total returns empty", func(t *testing.T) {
		query := &Query{Offset: 5}
		tasks, total, err := ScanTasksWithQuery([]string{tmpDir}, query)
		require.NoError(t, err)
		assert.Equal(t, 5, total)
		assert.Empty(t, tasks)
	})

	t.Run("offset plus limit exceeds total returns remainder", func(t *testing.T) {
		query := &Query{Offset: 3, Limit: 10}
		tasks, total, err := ScanTasksWithQuery([]string{tmpDir}, query)
		require.NoError(t, err)
		assert.Equal(t, 5, total)
		require.Len(t, tasks, 2)
		assert.Equal(t, "Task 4", tasks[0].Description)
		assert.Equal(t, "Task 5", tasks[1].Description)
	})

	t.Run("offset 0 limit 0 returns all", func(t *testing.T) {
		query := &Query{Offset: 0, Limit: 0}
		tasks, total, err := ScanTasksWithQuery([]string{tmpDir}, query)
		require.NoError(t, err)
		assert.Equal(t, 5, total)
		assert.Len(t, tasks, 5)
	})
}

func TestScanTasksWithQuerySortByPriority(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "tasks.md"), []byte(`# Tasks

- [ ] Low task 🔽
- [ ] High task ⏫
- [ ] No priority task
- [ ] Highest task 🔺
- [ ] Medium task 🔼
`), 0o600)
	require.NoError(t, err)

	query := &Query{
		SortBy: []SortKey{{Field: SortByPriority, Reverse: true}},
	}
	tasks, _, err := ScanTasksWithQuery([]string{tmpDir}, query)
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

	err := os.WriteFile(filepath.Join(tmpDir, "tasks.md"), []byte(`# Tasks

- [ ] March task 📅 2026-03-01
- [ ] January task 📅 2026-01-01
- [ ] No due task
- [ ] February task 📅 2026-02-01
`), 0o600)
	require.NoError(t, err)

	query := &Query{
		SortBy: []SortKey{{Field: SortByDue, Reverse: true}},
	}
	tasks, _, err := ScanTasksWithQuery([]string{tmpDir}, query)
	require.NoError(t, err)
	require.Len(t, tasks, 4)

	assert.Equal(t, "2026-03-01", tasks[0].DueDate)
	assert.Equal(t, "2026-02-01", tasks[1].DueDate)
	assert.Equal(t, "2026-01-01", tasks[2].DueDate)
	assert.Empty(t, tasks[3].DueDate)
}

func TestScanTasksWithQuerySortMultiKey(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "tasks.md"), []byte(`# Tasks

- [ ] High early ⏫ 📅 2026-01-01
- [ ] High late ⏫ 📅 2026-03-01
- [ ] Highest task 🔺 📅 2026-02-01
- [ ] High mid ⏫ 📅 2026-02-01
`), 0o600)
	require.NoError(t, err)

	query := &Query{
		SortBy: []SortKey{
			{Field: SortByPriority, Reverse: true},
			{Field: SortByDue, Reverse: true},
		},
	}
	tasks, _, err := ScanTasksWithQuery([]string{tmpDir}, query)
	require.NoError(t, err)
	require.Len(t, tasks, 4)

	assert.Equal(t, "Highest task", tasks[0].Description, "highest priority first")
	assert.Equal(t, "High late", tasks[1].Description, "same priority, latest due first")
	assert.Equal(t, "High mid", tasks[2].Description, "same priority, middle due")
	assert.Equal(t, "High early", tasks[3].Description, "same priority, earliest due last")
}
