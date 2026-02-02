package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:funlen // comprehensive test cases
func TestParseQuery(t *testing.T) {
	tests := []struct {
		check   func(*testing.T, *Query)
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "empty query",
			query:   "",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				assert.Empty(t, q.Filters)
			},
		},
		{
			name:    "status done",
			query:   "done",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				require.Len(t, q.Filters, 1)
				assert.IsType(t, &StatusFilter{}, q.Filters[0])
			},
		},
		{
			name:    "status not done",
			query:   "not done",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				require.Len(t, q.Filters, 1)
				assert.IsType(t, &StatusFilter{}, q.Filters[0])
			},
		},
		{
			name:    "due date on",
			query:   "due on 2024-01-15",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				require.Len(t, q.Filters, 1)
				assert.IsType(t, &DueDateFilter{}, q.Filters[0])
			},
		},
		{
			name:    "tag include",
			query:   "tag include #shopping",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				require.Len(t, q.Filters, 1)
				assert.IsType(t, &TagFilter{}, q.Filters[0])
			},
		},
		{
			name:    "multiple filters",
			query:   "not done\ntag include #shopping\ndue on or before 2024-01-15",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				assert.Len(t, q.Filters, 3)
			},
		},
		{
			name:    "ignore empty lines",
			query:   "done\n\nnot done",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				assert.Len(t, q.Filters, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQuery(tt.query)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestStatusFilter(t *testing.T) {
	tests := []struct {
		filter *StatusFilter
		task   *Task
		name   string
		want   bool
	}{
		{
			name:   "done filter matches complete task",
			filter: &StatusFilter{Done: true},
			task:   &Task{Status: "complete"},
			want:   true,
		},
		{
			name:   "done filter does not match incomplete task",
			filter: &StatusFilter{Done: true},
			task:   &Task{Status: "incomplete"},
			want:   false,
		},
		{
			name:   "not done filter matches incomplete task",
			filter: &StatusFilter{Done: false},
			task:   &Task{Status: "incomplete"},
			want:   true,
		},
		{
			name:   "not done filter does not match complete task",
			filter: &StatusFilter{Done: false},
			task:   &Task{Status: "complete"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Matches(tt.task)
			assert.Equal(t, tt.want, got)
		})
	}
}

//nolint:funlen // comprehensive test cases
func TestDueDateFilter(t *testing.T) {
	tests := []struct {
		filter *DueDateFilter
		task   *Task
		name   string
		want   bool
	}{
		{
			name:   "due on matches exact date",
			filter: &DueDateFilter{Op: DueOpOn, Date: "2024-01-15"},
			task:   &Task{DueDate: "2024-01-15"},
			want:   true,
		},
		{
			name:   "due on does not match different date",
			filter: &DueDateFilter{Op: DueOpOn, Date: "2024-01-15"},
			task:   &Task{DueDate: "2024-01-16"},
			want:   false,
		},
		{
			name:   "due on or before matches earlier date",
			filter: &DueDateFilter{Op: DueOpOnOrBefore, Date: "2024-01-15"},
			task:   &Task{DueDate: "2024-01-14"},
			want:   true,
		},
		{
			name:   "due on or before matches same date",
			filter: &DueDateFilter{Op: DueOpOnOrBefore, Date: "2024-01-15"},
			task:   &Task{DueDate: "2024-01-15"},
			want:   true,
		},
		{
			name:   "due on or before does not match later date",
			filter: &DueDateFilter{Op: DueOpOnOrBefore, Date: "2024-01-15"},
			task:   &Task{DueDate: "2024-01-16"},
			want:   false,
		},
		{
			name:   "due on or after matches later date",
			filter: &DueDateFilter{Op: DueOpOnOrAfter, Date: "2024-01-15"},
			task:   &Task{DueDate: "2024-01-16"},
			want:   true,
		},
		{
			name:   "due on or after matches same date",
			filter: &DueDateFilter{Op: DueOpOnOrAfter, Date: "2024-01-15"},
			task:   &Task{DueDate: "2024-01-15"},
			want:   true,
		},
		{
			name:   "due on or after does not match earlier date",
			filter: &DueDateFilter{Op: DueOpOnOrAfter, Date: "2024-01-15"},
			task:   &Task{DueDate: "2024-01-14"},
			want:   false,
		},
		{
			name:   "no due date matches task without due date",
			filter: &DueDateFilter{Op: DueOpNone},
			task:   &Task{DueDate: ""},
			want:   true,
		},
		{
			name:   "no due date does not match task with due date",
			filter: &DueDateFilter{Op: DueOpNone},
			task:   &Task{DueDate: "2024-01-15"},
			want:   false,
		},
		{
			name:   "has due date matches task with due date",
			filter: &DueDateFilter{Op: DueOpHas},
			task:   &Task{DueDate: "2024-01-15"},
			want:   true,
		},
		{
			name:   "has due date does not match task without due date",
			filter: &DueDateFilter{Op: DueOpHas},
			task:   &Task{DueDate: ""},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Matches(tt.task)
			assert.Equal(t, tt.want, got)
		})
	}
}

//nolint:funlen // comprehensive test cases
func TestTagFilter(t *testing.T) {
	tests := []struct {
		filter *TagFilter
		task   *Task
		name   string
		want   bool
	}{
		{
			name:   "tag include matches task with tag",
			filter: &TagFilter{Include: true, Tag: "shopping"},
			task:   &Task{Tags: []string{"shopping", "urgent"}},
			want:   true,
		},
		{
			name:   "tag include does not match task without tag",
			filter: &TagFilter{Include: true, Tag: "shopping"},
			task:   &Task{Tags: []string{"urgent"}},
			want:   false,
		},
		{
			name:   "tag do not include matches task without tag",
			filter: &TagFilter{Include: false, Tag: "shopping"},
			task:   &Task{Tags: []string{"urgent"}},
			want:   true,
		},
		{
			name:   "tag do not include does not match task with tag",
			filter: &TagFilter{Include: false, Tag: "shopping"},
			task:   &Task{Tags: []string{"shopping", "urgent"}},
			want:   false,
		},
		{
			name:   "has tags matches task with tags",
			filter: &TagFilter{HasAny: true},
			task:   &Task{Tags: []string{"shopping"}},
			want:   true,
		},
		{
			name:   "has tags does not match task without tags",
			filter: &TagFilter{HasAny: true},
			task:   &Task{Tags: []string{}},
			want:   false,
		},
		{
			name:   "no tags matches task without tags",
			filter: &TagFilter{HasAny: false},
			task:   &Task{Tags: []string{}},
			want:   true,
		},
		{
			name:   "no tags does not match task with tags",
			filter: &TagFilter{HasAny: false},
			task:   &Task{Tags: []string{"shopping"}},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Matches(tt.task)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPathFilter(t *testing.T) {
	tests := []struct {
		filter *PathFilter
		task   *Task
		name   string
		want   bool
	}{
		{
			name:   "path includes matches task in path",
			filter: &PathFilter{Include: true, Substring: "notes"},
			task:   &Task{FilePath: "notes/todo.md"},
			want:   true,
		},
		{
			name:   "path includes does not match task not in path",
			filter: &PathFilter{Include: true, Substring: "notes"},
			task:   &Task{FilePath: "other/todo.md"},
			want:   false,
		},
		{
			name:   "path does not include matches task not in path",
			filter: &PathFilter{Include: false, Substring: "notes"},
			task:   &Task{FilePath: "other/todo.md"},
			want:   true,
		},
		{
			name:   "path does not include does not match task in path",
			filter: &PathFilter{Include: false, Substring: "notes"},
			task:   &Task{FilePath: "notes/todo.md"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Matches(tt.task)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDescriptionFilter(t *testing.T) {
	tests := []struct {
		filter *DescriptionFilter
		task   *Task
		name   string
		want   bool
	}{
		{
			name:   "description includes matches task with substring",
			filter: &DescriptionFilter{Include: true, Substring: "groceries"},
			task:   &Task{Description: "Buy groceries"},
			want:   true,
		},
		{
			name:   "description includes does not match task without substring",
			filter: &DescriptionFilter{Include: true, Substring: "groceries"},
			task:   &Task{Description: "Buy milk"},
			want:   false,
		},
		{
			name:   "description includes is case sensitive",
			filter: &DescriptionFilter{Include: true, Substring: "Groceries"},
			task:   &Task{Description: "buy groceries"},
			want:   false,
		},
		{
			name:   "description does not include matches task without substring",
			filter: &DescriptionFilter{Include: false, Substring: "groceries"},
			task:   &Task{Description: "Buy milk"},
			want:   true,
		},
		{
			name:   "description does not include does not match task with substring",
			filter: &DescriptionFilter{Include: false, Substring: "groceries"},
			task:   &Task{Description: "Buy groceries"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Matches(tt.task)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestQueryMatches(t *testing.T) {
	query := &Query{
		Filters: []Filter{
			&StatusFilter{Done: false},
			&TagFilter{Include: true, Tag: "shopping"},
		},
	}

	tests := []struct {
		task *Task
		name string
		want bool
	}{
		{
			name: "matches all filters",
			task: &Task{
				Status: "incomplete",
				Tags:   []string{"shopping"},
			},
			want: true,
		},
		{
			name: "does not match status filter",
			task: &Task{
				Status: "complete",
				Tags:   []string{"shopping"},
			},
			want: false,
		},
		{
			name: "does not match tag filter",
			task: &Task{
				Status: "incomplete",
				Tags:   []string{"urgent"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := query.Matches(tt.task)
			assert.Equal(t, tt.want, got)
		})
	}
}
