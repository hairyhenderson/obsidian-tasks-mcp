package main

import (
	"testing"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
		check   func(*testing.T, *Query)
	}{
		{
			name:    "empty query",
			query:   "",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				if len(q.Filters) != 0 {
					t.Errorf("expected 0 filters, got %d", len(q.Filters))
				}
			},
		},
		{
			name:    "status done",
			query:   "done",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				if len(q.Filters) != 1 {
					t.Fatalf("expected 1 filter, got %d", len(q.Filters))
				}
				if _, ok := q.Filters[0].(*StatusFilter); !ok {
					t.Errorf("expected StatusFilter, got %T", q.Filters[0])
				}
			},
		},
		{
			name:    "status not done",
			query:   "not done",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				if len(q.Filters) != 1 {
					t.Fatalf("expected 1 filter, got %d", len(q.Filters))
				}
				if _, ok := q.Filters[0].(*StatusFilter); !ok {
					t.Errorf("expected StatusFilter, got %T", q.Filters[0])
				}
			},
		},
		{
			name:    "due date on",
			query:   "due on 2024-01-15",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				if len(q.Filters) != 1 {
					t.Fatalf("expected 1 filter, got %d", len(q.Filters))
				}
				if _, ok := q.Filters[0].(*DueDateFilter); !ok {
					t.Errorf("expected DueDateFilter, got %T", q.Filters[0])
				}
			},
		},
		{
			name:    "tag include",
			query:   "tag include #shopping",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				if len(q.Filters) != 1 {
					t.Fatalf("expected 1 filter, got %d", len(q.Filters))
				}
				if _, ok := q.Filters[0].(*TagFilter); !ok {
					t.Errorf("expected TagFilter, got %T", q.Filters[0])
				}
			},
		},
		{
			name:    "multiple filters",
			query:   "not done\ntag include #shopping\ndue on or before 2024-01-15",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				if len(q.Filters) != 3 {
					t.Errorf("expected 3 filters, got %d", len(q.Filters))
				}
			},
		},
		{
			name:    "ignore empty lines",
			query:   "done\n\nnot done",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				if len(q.Filters) != 2 {
					t.Errorf("expected 2 filters, got %d", len(q.Filters))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQuery(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestStatusFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter *StatusFilter
		task   *Task
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
			if got := tt.filter.Matches(tt.task); got != tt.want {
				t.Errorf("StatusFilter.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDueDateFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter *DueDateFilter
		task   *Task
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
			if got := tt.filter.Matches(tt.task); got != tt.want {
				t.Errorf("DueDateFilter.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter *TagFilter
		task   *Task
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
			if got := tt.filter.Matches(tt.task); got != tt.want {
				t.Errorf("TagFilter.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter *PathFilter
		task   *Task
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
			if got := tt.filter.Matches(tt.task); got != tt.want {
				t.Errorf("PathFilter.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDescriptionFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter *DescriptionFilter
		task   *Task
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
			if got := tt.filter.Matches(tt.task); got != tt.want {
				t.Errorf("DescriptionFilter.Matches() = %v, want %v", got, tt.want)
			}
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
		name string
		task *Task
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
			if got := query.Matches(tt.task); got != tt.want {
				t.Errorf("Query.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}
