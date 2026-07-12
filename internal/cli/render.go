package cli

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"

	"github.com/lohn/abwi/internal/ado"
	"github.com/lohn/abwi/internal/config"
)

// displayOrder lists the summary fields shown before the large text sections.
var displayOrder = []string{
	"System.State",
	"System.AssignedTo",
	"System.AreaPath",
	"System.IterationPath",
	"System.ChangedDate",
}

// fieldString renders a work item field value; identity fields arrive as
// objects with a displayName.
func fieldString(v any) string {
	if m, ok := v.(map[string]any); ok {
		if dn, ok := m["displayName"].(string); ok {
			return dn
		}
	}
	return fmt.Sprint(v)
}

func renderWorkItem(wi *workitemtracking.WorkItem, htmlFields map[string]bool) string {
	f := *wi.Fields
	var b strings.Builder
	fmt.Fprintf(&b, "#%d %s: %s\n", *wi.Id, f["System.WorkItemType"], f["System.Title"])
	for _, k := range displayOrder {
		if v, ok := f[k]; ok {
			fmt.Fprintf(&b, "%-14s %s\n", strings.TrimPrefix(k, "System.")+":", fieldString(v))
		}
	}
	var sections []string
	for k := range f {
		if htmlFields[k] {
			sections = append(sections, k)
		}
	}
	sort.Strings(sections)
	for _, k := range sections {
		fmt.Fprintf(&b, "\n## %s\n\n%s\n", k, fieldString(f[k]))
	}
	if wi.Relations != nil && len(*wi.Relations) > 0 {
		b.WriteString("\nRelations:\n")
		for _, r := range *wi.Relations {
			if r.Rel == nil || r.Url == nil {
				continue
			}
			fmt.Fprintf(&b, "  %s %s\n", ado.RelAlias(*r.Rel), *r.Url)
		}
	}
	return b.String()
}

func renderTable(items []workitemtracking.WorkItem) string {
	var b strings.Builder
	// The tabwriter only ever writes into the strings.Builder, so none of
	// these writes can fail.
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tTYPE\tSTATE\tTITLE")
	for _, wi := range items {
		f := *wi.Fields
		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", *wi.Id, f["System.WorkItemType"], f["System.State"], f["System.Title"])
	}
	_ = w.Flush()
	return b.String()
}

func renderComments(comments []ado.Comment) string {
	var b strings.Builder
	for i, c := range comments {
		if i > 0 {
			b.WriteString("\n---\n\n")
		}
		fmt.Fprintf(&b, "[%d] %s (%s)\n\n%s\n", c.ID, c.CreatedBy.DisplayName, c.CreatedDate, c.Text)
	}
	return b.String()
}

// printWorkItemLine prints the one-line result used by create and update.
func printWorkItemLine(c *ado.Client, wi *workitemtracking.WorkItem) {
	f := *wi.Fields
	fmt.Printf("#%d %s\n%s\n", *wi.Id, f["System.Title"], c.WebURL(*wi.Id))
}

func renderConfig(cfg *config.Config) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# global: %s\n", cfg.GlobalPath)
	local := cfg.LocalPath
	if local == "" {
		local = "(none)"
	}
	fmt.Fprintf(&b, "# local:  %s\n", local)
	for _, it := range cfg.Items() {
		fmt.Fprintf(&b, "%s = %q  # %s\n", it.Key, it.Value, it.Origin)
	}
	if len(cfg.Aliases) > 0 {
		keys := make([]string, 0, len(cfg.Aliases))
		for k := range cfg.Aliases {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		b.WriteString("\n[aliases]\n")
		for _, k := range keys {
			fmt.Fprintf(&b, "%s = %q\n", k, cfg.Aliases[k])
		}
	}
	return b.String()
}
