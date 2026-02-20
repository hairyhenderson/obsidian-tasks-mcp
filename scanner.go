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
//nolint:gocyclo // complexity from walking directories and filtering
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

	// Sort by file path then line number
	slices.SortFunc(allTasks, func(a, b *Task) int {
		if c := cmp.Compare(a.FilePath, b.FilePath); c != 0 {
			return c
		}

		return cmp.Compare(a.LineNumber, b.LineNumber)
	})

	return allTasks, nil
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
