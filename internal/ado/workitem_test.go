package ado

import (
	"testing"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
)

func TestHTMLFieldSet(t *testing.T) {
	fields := []workitemtracking.WorkItemField2{
		{ReferenceName: ptr("System.Description"), Type: &workitemtracking.FieldTypeValues.Html},
		{ReferenceName: ptr("System.Title"), Type: &workitemtracking.FieldTypeValues.String},
		{ReferenceName: ptr("Broken.NoType")},
	}
	got := htmlFieldSet(fields)
	if !got["System.Description"] || got["System.Title"] || len(got) != 1 {
		t.Errorf("htmlFieldSet = %v, want only System.Description", got)
	}
}

func TestWebURL(t *testing.T) {
	c := &Client{Org: "https://dev.azure.com/myorg", Project: "My Project"}
	want := "https://dev.azure.com/myorg/My Project/_workitems/edit/42"
	if got := c.WebURL(42); got != want {
		t.Errorf("WebURL = %q, want %q", got, want)
	}
}
