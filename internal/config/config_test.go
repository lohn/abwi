package config

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func clearEnv(t *testing.T) {
	t.Helper()
	t.Setenv("ABWI_ORG", "")
	t.Setenv("ABWI_PROJECT", "")
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)
	dir := t.TempDir()
	cfg, err := load(dir, filepath.Join(dir, "nope", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Format != "markdown" || cfg.Auth != "entra" {
		t.Errorf("got format=%q auth=%q, want markdown/entra", cfg.Format, cfg.Auth)
	}
	if cfg.Origins["format"] != "default" {
		t.Errorf("format origin = %q, want default", cfg.Origins["format"])
	}
	if cfg.LocalPath != "" {
		t.Errorf("LocalPath = %q, want empty", cfg.LocalPath)
	}
}

func TestLoadMergeAndOrigins(t *testing.T) {
	clearEnv(t)
	dir := t.TempDir()
	global := filepath.Join(dir, "g", "config.toml")
	write(t, global, `
org = "https://dev.azure.com/gorg"
project = "GProj"
format = "html"
[aliases]
ac = "Global.AC"
repro = "Global.Repro"
`)
	repo := filepath.Join(dir, "repo")
	write(t, filepath.Join(repo, ".abwi.toml"), `
project = "LProj"
[aliases]
ac = "Local.AC"
`)
	sub := filepath.Join(repo, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg, err := load(sub, global)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][2]string{
		"org":     {"https://dev.azure.com/gorg", "global"},
		"project": {"LProj", "local"},
		"format":  {"html", "global"},
		"auth":    {"entra", "default"},
	}
	for _, it := range cfg.Items() {
		w, ok := want[it.Key]
		if !ok {
			continue
		}
		if it.Value != w[0] || it.Origin != w[1] {
			t.Errorf("%s: got (%q, %q), want (%q, %q)", it.Key, it.Value, it.Origin, w[0], w[1])
		}
	}
	if cfg.Aliases["ac"] != "Local.AC" || cfg.Aliases["repro"] != "Global.Repro" {
		t.Errorf("aliases merged wrong: %v", cfg.Aliases)
	}
	if cfg.LocalPath != filepath.Join(repo, ".abwi.toml") {
		t.Errorf("LocalPath = %q", cfg.LocalPath)
	}
}

func TestEnvOverride(t *testing.T) {
	clearEnv(t)
	t.Setenv("ABWI_ORG", "https://dev.azure.com/eorg")
	dir := t.TempDir()
	cfg, err := load(dir, filepath.Join(dir, "nope.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Org != "https://dev.azure.com/eorg" || cfg.Origins["org"] != "env" {
		t.Errorf("got org=%q origin=%q, want env override", cfg.Org, cfg.Origins["org"])
	}
}

func TestLocalAuthIgnored(t *testing.T) {
	clearEnv(t)
	dir := t.TempDir()
	write(t, filepath.Join(dir, ".abwi.toml"), "auth = \"pat\"\n")
	cfg, err := load(dir, filepath.Join(dir, "nope.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Auth != "entra" || cfg.Origins["auth"] != "default" {
		t.Errorf("got auth=%q origin=%q, want the local value ignored (entra/default)", cfg.Auth, cfg.Origins["auth"])
	}
}

func TestGlobalAuth(t *testing.T) {
	clearEnv(t)
	dir := t.TempDir()
	global := filepath.Join(dir, "g", "config.toml")
	write(t, global, "auth = \"pat\"\n")
	cfg, err := load(dir, global)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Auth != "pat" || cfg.Origins["auth"] != "global" {
		t.Errorf("got auth=%q origin=%q, want pat/global", cfg.Auth, cfg.Origins["auth"])
	}
}

func TestInvalidFormat(t *testing.T) {
	clearEnv(t)
	dir := t.TempDir()
	write(t, filepath.Join(dir, ".abwi.toml"), "format = \"rst\"\n")
	if _, err := load(dir, filepath.Join(dir, "nope.toml")); err == nil {
		t.Fatal("want error for invalid format")
	}
}
