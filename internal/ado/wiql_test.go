package ado

import (
	"strings"
	"testing"
)

func TestBuildWIQL(t *testing.T) {
	cases := []struct {
		name                      string
		typ, state, assignee      string
		all                       bool
		wantContains, wantMissing []string
	}{
		{
			name:         "default is mine",
			wantContains: []string{"[System.AssignedTo] = @Me", "[System.TeamProject] = @project", "ORDER BY [System.ChangedDate] DESC"},
		},
		{
			name:        "all drops assignee",
			all:         true,
			wantMissing: []string{"System.AssignedTo"},
		},
		{
			name:         "filters and escaping",
			typ:          "Bug",
			state:        "Active",
			assignee:     "O'Brien",
			wantContains: []string{"[System.WorkItemType] = 'Bug'", "[System.State] = 'Active'", "[System.AssignedTo] = 'O''Brien'"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := BuildWIQL(c.typ, c.state, c.assignee, c.all)
			for _, w := range c.wantContains {
				if !strings.Contains(got, w) {
					t.Errorf("missing %q in %q", w, got)
				}
			}
			for _, w := range c.wantMissing {
				if strings.Contains(got, w) {
					t.Errorf("unexpected %q in %q", w, got)
				}
			}
		})
	}
}
