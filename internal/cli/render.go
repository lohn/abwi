package cli

import (
	"fmt"
	"sort"
	"strings"

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
