package ado

import "testing"

func TestCommentsURL(t *testing.T) {
	got := commentsURL("https://dev.azure.com/org/", "My Project", 42, "markdown")
	want := "https://dev.azure.com/org/My%20Project/_apis/wit/workItems/42/comments?api-version=7.1-preview.4&format=markdown"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	got = commentsURL("https://dev.azure.com/org", "p", 7, "")
	want = "https://dev.azure.com/org/p/_apis/wit/workItems/7/comments?api-version=7.1-preview.4"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
