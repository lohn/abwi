package ado

import (
	"reflect"
	"testing"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
)

func TestBuildFieldPatchMarkdown(t *testing.T) {
	ops, err := BuildFieldPatch(
		map[string]string{"System.Title": "T", "System.Description": "# d"},
		map[string]bool{"System.Description": true},
		FormatMarkdown,
	)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, op := range ops {
		got = append(got, *op.Path)
	}
	want := []string{
		"/fields/System.Description",
		"/multilineFieldsFormat/System.Description",
		"/fields/System.Title",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("paths = %v, want %v", got, want)
	}
	if ops[1].Value != "Markdown" {
		t.Errorf("multilineFieldsFormat value = %v, want Markdown", ops[1].Value)
	}
}

func TestBuildFieldPatchHTML(t *testing.T) {
	ops, err := BuildFieldPatch(
		map[string]string{"System.Description": "# d"},
		map[string]bool{"System.Description": true},
		FormatHTML,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(ops) != 1 {
		t.Fatalf("got %d ops, want 1 (no multilineFieldsFormat in html mode)", len(ops))
	}
	if ops[0].Value != "<h1>d</h1>\n" {
		t.Errorf("value = %q, want converted HTML", ops[0].Value)
	}
}

func TestLinkTypeRef(t *testing.T) {
	cases := []struct{ in, want string }{
		{"parent", "System.LinkTypes.Hierarchy-Reverse"},
		{"child", "System.LinkTypes.Hierarchy-Forward"},
		{"related", "System.LinkTypes.Related"},
		{"", "System.LinkTypes.Related"},
		{"System.LinkTypes.Duplicate-Forward", "System.LinkTypes.Duplicate-Forward"},
	}
	for _, c := range cases {
		got, err := LinkTypeRef(c.in)
		if err != nil || got != c.want {
			t.Errorf("LinkTypeRef(%q) = (%q, %v), want %q", c.in, got, err, c.want)
		}
	}
	if _, err := LinkTypeRef("bogus"); err == nil {
		t.Error("want error for unknown alias without a dot")
	}
}

func TestFindRelationIndexes(t *testing.T) {
	rel := func(typeRef, id string) workitemtracking.WorkItemRelation {
		return workitemtracking.WorkItemRelation{
			Rel: ptr(typeRef),
			Url: ptr("https://dev.azure.com/o/_apis/wit/workItems/" + id),
		}
	}
	rels := []workitemtracking.WorkItemRelation{
		rel("System.LinkTypes.Related", "12"),
		rel("System.LinkTypes.Hierarchy-Reverse", "12"),
		rel("System.LinkTypes.Related", "123"),
	}
	if got := FindRelationIndexes(rels, 12, ""); !reflect.DeepEqual(got, []int{0, 1}) {
		t.Errorf("all types: got %v, want [0 1]", got)
	}
	if got := FindRelationIndexes(rels, 12, "System.LinkTypes.Related"); !reflect.DeepEqual(got, []int{0}) {
		t.Errorf("filtered: got %v, want [0]", got)
	}
	if got := FindRelationIndexes(rels, 99, ""); got != nil {
		t.Errorf("no match: got %v, want nil", got)
	}
}

func TestRemoveRelationOps(t *testing.T) {
	ops := RemoveRelationOps([]int{1, 5, 3})
	var got []string
	for _, op := range ops {
		got = append(got, *op.Path)
	}
	want := []string{"/relations/5", "/relations/3", "/relations/1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("paths = %v, want %v (descending)", got, want)
	}
}
