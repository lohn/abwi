// Package ado wraps the Azure DevOps SDK for work item operations.
package ado

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
	"github.com/yuin/goldmark"
)

const (
	FormatMarkdown = "markdown"
	FormatHTML     = "html"

	relParent  = "System.LinkTypes.Hierarchy-Reverse"
	relChild   = "System.LinkTypes.Hierarchy-Forward"
	relRelated = "System.LinkTypes.Related"
)

func ptr[T any](v T) *T { return &v }

// BuildFieldPatch turns field values into JSON Patch operations, ordered by
// field name. For fields in htmlFields (the project's large-text fields),
// format "markdown" adds a multilineFieldsFormat op so Azure Boards stores the
// value as native Markdown; format "html" converts the value from Markdown to
// HTML instead (for organizations without Markdown support).
func BuildFieldPatch(fields map[string]string, htmlFields map[string]bool, format string) ([]webapi.JsonPatchOperation, error) {
	names := make([]string, 0, len(fields))
	for n := range fields {
		names = append(names, n)
	}
	sort.Strings(names)
	ops := make([]webapi.JsonPatchOperation, 0, len(names))
	for _, n := range names {
		value := fields[n]
		if htmlFields[n] && format == FormatHTML {
			var buf bytes.Buffer
			if err := goldmark.Convert([]byte(value), &buf); err != nil {
				return nil, fmt.Errorf("converting %s to HTML: %w", n, err)
			}
			value = buf.String()
		}
		ops = append(ops, addOp("/fields/"+n, value))
		if htmlFields[n] && format == FormatMarkdown {
			ops = append(ops, addOp("/multilineFieldsFormat/"+n, "Markdown"))
		}
	}
	return ops, nil
}

func addOp(path string, value any) webapi.JsonPatchOperation {
	return webapi.JsonPatchOperation{
		Op:    &webapi.OperationValues.Add,
		Path:  ptr(path),
		Value: value,
	}
}

// RelationPatch returns the op that adds a relation of type typeRef to targetID.
func RelationPatch(orgURL, typeRef string, targetID int) webapi.JsonPatchOperation {
	return addOp("/relations/-", map[string]any{
		"rel": typeRef,
		"url": fmt.Sprintf("%s/_apis/wit/workItems/%d", strings.TrimRight(orgURL, "/"), targetID),
	})
}

// LinkTypeRef resolves a CLI link type (parent, child, related, or a raw
// reference name containing a dot) to a relation reference name. An empty
// name defaults to related.
func LinkTypeRef(name string) (string, error) {
	switch name {
	case "parent":
		return relParent, nil
	case "child":
		return relChild, nil
	case "related", "":
		return relRelated, nil
	}
	if strings.Contains(name, ".") {
		return name, nil
	}
	return "", fmt.Errorf("unknown link type %q: use parent, child, related, or a full reference name", name)
}

// RelAlias is the inverse of LinkTypeRef, for display.
func RelAlias(rel string) string {
	switch rel {
	case relParent:
		return "parent"
	case relChild:
		return "child"
	case relRelated:
		return "related"
	}
	return rel
}

// FindRelationIndexes returns the indexes of relations pointing at targetID,
// filtered by typeRef when non-empty.
func FindRelationIndexes(relations []workitemtracking.WorkItemRelation, targetID int, typeRef string) []int {
	suffix := fmt.Sprintf("/workitems/%d", targetID)
	var idx []int
	for i, r := range relations {
		if r.Url == nil || !strings.HasSuffix(strings.ToLower(*r.Url), suffix) {
			continue
		}
		if typeRef != "" && (r.Rel == nil || *r.Rel != typeRef) {
			continue
		}
		idx = append(idx, i)
	}
	return idx
}

// RemoveRelationOps returns remove ops ordered from the highest index down,
// so earlier indexes remain valid while the server applies the patch in order.
func RemoveRelationOps(indexes []int) []webapi.JsonPatchOperation {
	sorted := append([]int(nil), indexes...)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	ops := make([]webapi.JsonPatchOperation, 0, len(sorted))
	for _, i := range sorted {
		ops = append(ops, webapi.JsonPatchOperation{
			Op:   &webapi.OperationValues.Remove,
			Path: ptr(fmt.Sprintf("/relations/%d", i)),
		})
	}
	return ops
}
