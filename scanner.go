package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ScanTasks scans markdown files in the given root directories and returns all tasks
func ScanTasks(roots []string) ([]*Task, error) {
	return ScanTasksWithQuery(roots, nil)
}

// ScanTasksWithQuery scans markdown files and filters tasks using the provided query
func ScanTasksWithQuery(roots []string, query *Query) ([]*Task, error) {
	var allTasks []*Task
	
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
			
			tasks, err := parseTasksFromFile(path, absRoot)
			if err != nil {
				// Log error but continue scanning other files
				return nil
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
	sort.Slice(allTasks, func(i, j int) bool {
		if allTasks[i].FilePath != allTasks[j].FilePath {
			return allTasks[i].FilePath < allTasks[j].FilePath
		}
		return allTasks[i].LineNumber < allTasks[j].LineNumber
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
