package main

import (
	"bufio"
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// ScanTasks scans markdown files in the given root directories and returns all tasks
func ScanTasks(roots []string) ([]*Task, error) {
	return ScanTasksWithQuery(roots, nil)
}

// ScanTasksWithQuery scans markdown files and filters tasks using the provided query
//
//nolint:gocyclo,gocognit,funlen,nestif // complexity from walking directories and filtering
func ScanTasksWithQuery(roots []string, query *Query) ([]*Task, error) {
	allTasks := make([]*Task, 0)

	for _, root := range roots {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for %q: %w", root, err)
		}

		err = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Only process markdown files
			if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
				return nil
			}

			tasks, parseErr := parseTasksFromFile(path, absRoot)
			if parseErr != nil {
				// Log error but continue scanning other files
				// We could log parseErr here if we had a logger
				return nil //nolint:nilerr // intentionally continue on parse errors
			}

			// Filter tasks if query is provided
			if query != nil {
				for _, task := range tasks {
					if query.Matches(task) {
						allTasks = append(allTasks, task)
					}
				}
			} else {
				allTasks = append(allTasks, tasks...)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk directory %q: %w", root, err)
		}
	}

	if query != nil && len(query.SortBy) > 0 {
		slices.SortStableFunc(allTasks, func(a, b *Task) int {
			for _, key := range query.SortBy {
				sentinel, c := compareByField(a, b, key.Field)
				if sentinel != 0 {
					return sentinel
				}

				if c != 0 {
					if key.Reverse {
						return -c
					}

					return c
				}
			}

			if c := cmp.Compare(a.FilePath, b.FilePath); c != 0 {
				return c
			}

			return cmp.Compare(a.LineNumber, b.LineNumber)
		})
	} else {
		slices.SortFunc(allTasks, func(a, b *Task) int {
			if c := cmp.Compare(a.FilePath, b.FilePath); c != 0 {
				return c
			}

			return cmp.Compare(a.LineNumber, b.LineNumber)
		})
	}

	if query != nil && query.Limit > 0 && len(allTasks) > query.Limit {
		allTasks = allTasks[:query.Limit]
	}

	return allTasks, nil
}

// compareByField returns (sentinel, cmp). If sentinel != 0 it takes
// priority and must not be reversed (used for "always sort last" values
// like PriorityNone / empty due date). Otherwise cmp is the normal
// comparison result that may be reversed by the caller.
//
//nolint:gocyclo // branching per field + sentinel cases
func compareByField(a, b *Task, field SortField) (int, int) {
	switch field {
	case SortByPriority:
		switch {
		case a.Priority == PriorityNone && b.Priority == PriorityNone:
			return 0, 0
		case a.Priority == PriorityNone:
			return 1, 0
		case b.Priority == PriorityNone:
			return -1, 0
		}

		return 0, cmp.Compare(a.Priority, b.Priority)
	case SortByDue:
		switch {
		case a.DueDate == "" && b.DueDate == "":
			return 0, 0
		case a.DueDate == "":
			return 1, 0
		case b.DueDate == "":
			return -1, 0
		}

		return 0, compareDates(a.DueDate, b.DueDate)
	default:
		return 0, 0
	}
}

func parseTasksFromFile(filePath, rootDir string) ([]*Task, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tasks []*Task

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		task := ParseTask(line, filePath, lineNumber)
		if task != nil {
			// Make file path relative to root if possible
			relPath, err := filepath.Rel(rootDir, filePath)
			if err == nil {
				task.FilePath = relPath
			}

			tasks = append(tasks, task)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %q: %w", filePath, err)
	}

	return tasks, nil
}
