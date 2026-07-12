package ado

import (
	"fmt"
	"strings"
)

// BuildWIQL assembles the default list query. With no assignee and all=false
// the query is scoped to @Me.
func BuildWIQL(workItemType, state, assignee string, all bool) string {
	conds := []string{"[System.TeamProject] = @project"}
	switch {
	case assignee != "":
		conds = append(conds, fmt.Sprintf("[System.AssignedTo] = '%s'", escapeWIQL(assignee)))
	case !all:
		conds = append(conds, "[System.AssignedTo] = @Me")
	}
	if workItemType != "" {
		conds = append(conds, fmt.Sprintf("[System.WorkItemType] = '%s'", escapeWIQL(workItemType)))
	}
	if state != "" {
		conds = append(conds, fmt.Sprintf("[System.State] = '%s'", escapeWIQL(state)))
	}
	return "SELECT [System.Id] FROM WorkItems WHERE " +
		strings.Join(conds, " AND ") + " ORDER BY [System.ChangedDate] DESC"
}

// escapeWIQL doubles single quotes for WIQL string literals.
func escapeWIQL(s string) string { return strings.ReplaceAll(s, "'", "''") }
