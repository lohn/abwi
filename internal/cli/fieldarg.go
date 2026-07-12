package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// expandArg resolves curl-style file references: "@path" reads the file,
// "@-" reads stdin, and a leading `\@` escapes a literal "@".
func expandArg(s string, stdin io.Reader) (string, error) {
	switch {
	case strings.HasPrefix(s, `\@`):
		return s[1:], nil
	case s == "@-":
		b, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return string(b), nil
	case strings.HasPrefix(s, "@"):
		b, err := os.ReadFile(s[1:])
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return s, nil
}

// parseFields parses repeated "name=value" flags, resolving aliases and @
// references in values.
func parseFields(args []string, aliases map[string]string, stdin io.Reader) (map[string]string, error) {
	fields := make(map[string]string, len(args))
	for _, a := range args {
		name, value, ok := strings.Cut(a, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --field %q: expected <name>=<value>", a)
		}
		if ref, ok := aliases[name]; ok {
			name = ref
		}
		v, err := expandArg(value, stdin)
		if err != nil {
			return nil, fmt.Errorf("--field %s: %w", name, err)
		}
		fields[name] = v
	}
	return fields, nil
}
