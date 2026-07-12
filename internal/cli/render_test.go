package cli

import (
	"strings"
	"testing"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
)

func strPtr(s string) *string { return &s }

func TestRenderWorkItem(t *testing.T) {
	fields := map[string]any{
		"System.WorkItemType": "Bug",
		"System.Title":        "It breaks",
		"System.State":        "Active",
		"System.AssignedTo":   map[string]any{"displayName": "Lohn IMAI"},
		"System.Description":  "# repro\n\nsteps",
	}
	id := 42
	rels := []workitemtracking.WorkItemRelation{{
		Rel: strPtr("System.LinkTypes.Hierarchy-Reverse"),
		Url: strPtr("https://dev.azure.com/o/_apis/wit/workItems/7"),
	}}
	wi := &workitemtracking.WorkItem{Id: &id, Fields: &fields, Relations: &rels}
	got := renderWorkItem(wi, map[string]bool{"System.Description": true})
	for _, want := range []string{
		"#42 Bug: It breaks",
		"State:",
		"Active",
		"Lohn IMAI",
		"## System.Description",
		"# repro",
		"parent https://dev.azure.com/o/_apis/wit/workItems/7",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q:\n%s", want, got)
		}
	}
}
