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
	tasks, _, err := ScanTasksWithQuery(roots, nil)

	return tasks, err
}

// ScanTasksWithQuery scans markdown files and filters tasks using the provided query
//
//nolint:gocyclo // complexity from walking directories and filtering
func ScanTasksWithQuery(roots []string, query *Query) ([]*Task, int, error) {
	allTasks := make([]*Task, 0)

	for _, root := range roots {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get absolute path for %q: %w", root, err)
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
			return nil, 0, fmt.Errorf("failed to walk directory %q: %w", root, err)
		}
	}

	sortTasks(allTasks, query)

	total := len(allTasks)
	allTasks = applyPagination(allTasks, query)

	return allTasks, total, nil
}

//nolint:gocognit,gocyclo // multi-key sort with special "none/empty sorts last" logic
func sortTasks(tasks []*Task, query *Query) {
	slices.SortStableFunc(tasks, func(a, b *Task) int {
		if query != nil {
			for _, key := range query.SortBy {
				var c int

				switch key.Field {
				case SortByPriority:
					switch {
					case a.Priority == PriorityNone && b.Priority == PriorityNone:
						c = 0
					case a.Priority == PriorityNone:
						return 1
					case b.Priority == PriorityNone:
						return -1
					default:
						c = cmp.Compare(a.Priority, b.Priority)
					}
				case SortByDue:
					switch {
					case a.DueDate == "" && b.DueDate == "":
						c = 0
					case a.DueDate == "":
						return 1
					case b.DueDate == "":
						return -1
					default:
						c = cmp.Compare(a.DueDate, b.DueDate)
					}
				}

				if c != 0 {
					if key.Reverse {
						return -c
					}

					return c
				}
			}
		}

		if c := cmp.Compare(a.FilePath, b.FilePath); c != 0 {
			return c
		}

		return cmp.Compare(a.LineNumber, b.LineNumber)
	})
}

func applyPagination(tasks []*Task, query *Query) []*Task {
	if query == nil {
		return tasks
	}

	if query.Offset > 0 {
		if query.Offset >= len(tasks) {
			return tasks[:0]
		}

		tasks = tasks[query.Offset:]
	}

	if query.Limit > 0 && query.Limit < len(tasks) {
		tasks = tasks[:query.Limit]
	}

	return tasks
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
