package cli

import (
	"strings"
	"testing"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"

	"github.com/lohn/abwi/internal/config"
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

func TestRenderTable(t *testing.T) {
	mk := func(id int, typ, state, title string) workitemtracking.WorkItem {
		f := map[string]any{"System.WorkItemType": typ, "System.State": state, "System.Title": title}
		return workitemtracking.WorkItem{Id: &id, Fields: &f}
	}
	got := renderTable([]workitemtracking.WorkItem{
		mk(1, "Bug", "Active", "Crash on save"),
		mk(23, "Task", "New", "Write docs"),
	})
	if !strings.Contains(got, "ID") || !strings.Contains(got, "Crash on save") || !strings.Contains(got, "23") {
		t.Errorf("table output wrong:\n%s", got)
	}
}

func TestRenderConfig(t *testing.T) {
	cfg := &config.Config{
		Org: "https://dev.azure.com/o", Project: "P", Format: "markdown", Auth: "entra",
		Aliases:    map[string]string{"ac": "Microsoft.VSTS.Common.AcceptanceCriteria"},
		Origins:    map[string]string{"org": "global", "project": "local", "format": "default", "auth": "default", "default-type": "default"},
		GlobalPath: "/home/u/.config/abwi/config.toml",
	}
	got := renderConfig(cfg)
	for _, want := range []string{
		"# global: /home/u/.config/abwi/config.toml",
		"# local:  (none)",
		`org = "https://dev.azure.com/o"  # global`,
		`project = "P"  # local`,
		`format = "markdown"  # default`,
		"[aliases]",
		`ac = "Microsoft.VSTS.Common.AcceptanceCriteria"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q:\n%s", want, got)
		}
	}
}
