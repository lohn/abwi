package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandArg(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "d.md")
	if err := os.WriteFile(file, []byte("# hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, in, want string }{
		{"plain", "hello", "hello"},
		{"file", "@" + file, "# hi\n"},
		{"stdin", "@-", "from stdin"},
		{"escaped", `\@literal`, "@literal"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := expandArg(c.in, strings.NewReader("from stdin"))
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("expandArg(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
	if _, err := expandArg("@"+filepath.Join(dir, "missing.md"), nil); err == nil {
		t.Error("want error for missing file")
	}
}

func TestParseFields(t *testing.T) {
	aliases := map[string]string{"ac": "Microsoft.VSTS.Common.AcceptanceCriteria"}
	got, err := parseFields(
		[]string{"System.Title=T", "ac=done when green"},
		aliases, strings.NewReader(""),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got["System.Title"] != "T" {
		t.Errorf("System.Title = %q", got["System.Title"])
	}
	if got["Microsoft.VSTS.Common.AcceptanceCriteria"] != "done when green" {
		t.Errorf("alias not resolved: %v", got)
	}
	if _, err := parseFields([]string{"no-equals"}, nil, nil); err == nil {
		t.Error("want error for missing '='")
	}
}
