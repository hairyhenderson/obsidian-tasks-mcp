package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Filter is an interface for task filters
type Filter interface {
	Matches(task *Task) bool
}

// Query represents a parsed query with filters
type Query struct {
	Filters []Filter
}

// StatusFilter filters tasks by completion status
type StatusFilter struct {
	Done bool
}

func (f *StatusFilter) Matches(task *Task) bool {
	if f.Done {
		return task.Status == "complete"
	}

	return task.Status == "incomplete"
}

// DueDateOp represents a due date comparison operation
type DueDateOp int

const (
	DueOpOn DueDateOp = iota
	DueOpOnOrBefore
	DueOpOnOrAfter
	DueOpNone
	DueOpHas
)

// DueDateFilter filters tasks by due date
type DueDateFilter struct {
	Date string
	Op   DueDateOp
}

func (f *DueDateFilter) Matches(task *Task) bool {
	switch f.Op {
	case DueOpNone:
		return task.DueDate == ""
	case DueOpHas:
		return task.DueDate != ""
	case DueOpOn:
		return task.DueDate == f.Date
	case DueOpOnOrBefore:
		if task.DueDate == "" {
			return false
		}

		return compareDates(task.DueDate, f.Date) <= 0
	case DueOpOnOrAfter:
		if task.DueDate == "" {
			return false
		}

		return compareDates(task.DueDate, f.Date) >= 0
	default:
		return false
	}
}

func compareDates(date1, date2 string) int {
	t1, err1 := time.Parse("2006-01-02", date1)

	t2, err2 := time.Parse("2006-01-02", date2)
	if err1 != nil || err2 != nil {
		return 0
	}

	if t1.Before(t2) {
		return -1
	}

	if t1.After(t2) {
		return 1
	}

	return 0
}

// TagFilter filters tasks by tags
type TagFilter struct {
	Tag     string
	Include bool
	HasAny  bool
}

func (f *TagFilter) Matches(task *Task) bool {
	// If Tag is empty, we're checking for "has tags" or "no tags"
	if f.Tag == "" {
		hasTags := len(task.Tags) > 0
		// HasAny: true means "has tags", false means "no tags"
		if f.HasAny {
			return hasTags
		}

		return !hasTags
	}

	// "tag include #tag" or "tag do not include #tag"
	for _, tag := range task.Tags {
		if tag == f.Tag {
			return f.Include
		}
	}

	return !f.Include
}

// PathFilter filters tasks by file path
type PathFilter struct {
	Substring string
	Include   bool
}

func (f *PathFilter) Matches(task *Task) bool {
	contains := strings.Contains(task.FilePath, f.Substring)
	if f.Include {
		return contains
	}

	return !contains
}

// DescriptionFilter filters tasks by description
type DescriptionFilter struct {
	Substring string
	Include   bool
}

func (f *DescriptionFilter) Matches(task *Task) bool {
	contains := strings.Contains(task.Description, f.Substring)
	if f.Include {
		return contains
	}

	return !contains
}

var (
	statusDoneRegex    = regexp.MustCompile(`^done$`)
	statusNotDoneRegex = regexp.MustCompile(`^not done$`)

	dueOnRegex         = regexp.MustCompile(`^due on (\d{4}-\d{2}-\d{2})$`)
	dueOnOrBeforeRegex = regexp.MustCompile(`^due on or before (\d{4}-\d{2}-\d{2})$`)
	dueOnOrAfterRegex  = regexp.MustCompile(`^due on or after (\d{4}-\d{2}-\d{2})$`)
	dueNoneRegex       = regexp.MustCompile(`^no due date$`)
	dueHasRegex        = regexp.MustCompile(`^has due date$`)

	tagIncludeRegex    = regexp.MustCompile(`^tag include #(\w+)$`)
	tagNotIncludeRegex = regexp.MustCompile(`^tag do not include #(\w+)$`)
	tagHasRegex        = regexp.MustCompile(`^has tags$`)
	tagNoRegex         = regexp.MustCompile(`^no tags$`)

	pathIncludesRegex    = regexp.MustCompile(`^path includes (.+)$`)
	pathNotIncludesRegex = regexp.MustCompile(`^path does not include (.+)$`)

	descIncludesRegex    = regexp.MustCompile(`^description includes (.+)$`)
	descNotIncludesRegex = regexp.MustCompile(`^description does not include (.+)$`)
)

// ParseQuery parses a query string into a Query struct
func ParseQuery(queryStr string) (*Query, error) {
	query := &Query{Filters: []Filter{}}

	lines := strings.Split(queryStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip comments (lines starting with #)
		if strings.HasPrefix(line, "#") {
			continue
		}

		filter, err := parseFilterLine(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse filter line %q: %w", line, err)
		}

		if filter != nil {
			query.Filters = append(query.Filters, filter)
		}
	}

	return query, nil
}

//nolint:gocyclo,funlen,unparam // parsing different filter types requires branching
func parseFilterLine(line string) (Filter, error) {
	line = strings.TrimSpace(line)

	// Status filters
	if statusDoneRegex.MatchString(line) {
		return &StatusFilter{Done: true}, nil
	}

	if statusNotDoneRegex.MatchString(line) {
		return &StatusFilter{Done: false}, nil
	}

	// Due date filters
	if matches := dueOnRegex.FindStringSubmatch(line); len(matches) >= 2 {
		return &DueDateFilter{Op: DueOpOn, Date: matches[1]}, nil
	}

	if matches := dueOnOrBeforeRegex.FindStringSubmatch(line); len(matches) >= 2 {
		return &DueDateFilter{Op: DueOpOnOrBefore, Date: matches[1]}, nil
	}

	if matches := dueOnOrAfterRegex.FindStringSubmatch(line); len(matches) >= 2 {
		return &DueDateFilter{Op: DueOpOnOrAfter, Date: matches[1]}, nil
	}

	if dueNoneRegex.MatchString(line) {
		return &DueDateFilter{Op: DueOpNone}, nil
	}

	if dueHasRegex.MatchString(line) {
		return &DueDateFilter{Op: DueOpHas}, nil
	}

	// Tag filters
	if matches := tagIncludeRegex.FindStringSubmatch(line); len(matches) >= 2 {
		return &TagFilter{Include: true, Tag: matches[1]}, nil
	}

	if matches := tagNotIncludeRegex.FindStringSubmatch(line); len(matches) >= 2 {
		return &TagFilter{Include: false, Tag: matches[1]}, nil
	}

	if tagHasRegex.MatchString(line) {
		return &TagFilter{HasAny: true}, nil
	}

	if tagNoRegex.MatchString(line) {
		return &TagFilter{HasAny: false}, nil
	}

	// Path filters
	if matches := pathIncludesRegex.FindStringSubmatch(line); len(matches) >= 2 {
		return &PathFilter{Include: true, Substring: matches[1]}, nil
	}

	if matches := pathNotIncludesRegex.FindStringSubmatch(line); len(matches) >= 2 {
		return &PathFilter{Include: false, Substring: matches[1]}, nil
	}

	// Description filters
	if matches := descIncludesRegex.FindStringSubmatch(line); len(matches) >= 2 {
		return &DescriptionFilter{Include: true, Substring: matches[1]}, nil
	}

	if matches := descNotIncludesRegex.FindStringSubmatch(line); len(matches) >= 2 {
		return &DescriptionFilter{Include: false, Substring: matches[1]}, nil
	}

	// Unknown filter - return nil to skip
	return nil, nil
}

// Matches checks if a task matches all filters in the query
func (q *Query) Matches(task *Task) bool {
	for _, filter := range q.Filters {
		if !filter.Matches(task) {
			return false
		}
	}

	return true
}
