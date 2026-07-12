// Package config loads and merges abwi configuration from TOML files and
// environment variables, tracking where each value came from.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const localName = ".abwi.toml"

// Config is the merged configuration with per-key origin tracking.
type Config struct {
	Org         string
	Project     string
	Format      string // "markdown" or "html"
	Auth        string // "entra" or "pat"
	DefaultType string
	Aliases     map[string]string
	Origins     map[string]string // key -> "default" | "global" | "local" | "env" | "flag"
	GlobalPath  string
	LocalPath   string // "" when no local config was found
}

// Item is one resolved scalar entry, for display.
type Item struct{ Key, Value, Origin string }

// Items returns the scalar entries in stable display order.
func (c *Config) Items() []Item {
	return []Item{
		{"org", c.Org, c.Origins["org"]},
		{"project", c.Project, c.Origins["project"]},
		{"format", c.Format, c.Origins["format"]},
		{"auth", c.Auth, c.Origins["auth"]},
		{"default-type", c.DefaultType, c.Origins["default-type"]},
	}
}

type fileConfig struct {
	Org         string            `toml:"org"`
	Project     string            `toml:"project"`
	Format      string            `toml:"format"`
	Auth        string            `toml:"auth"`
	DefaultType string            `toml:"default-type"`
	Aliases     map[string]string `toml:"aliases"`
}

// Load reads the global config, the nearest .abwi.toml at or above cwd, and
// ABWI_* environment variables, in increasing precedence.
func Load(cwd string) (*Config, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return load(cwd, filepath.Join(dir, "abwi", "config.toml"))
}

func load(cwd, globalPath string) (*Config, error) {
	cfg := &Config{
		Format:  "markdown",
		Auth:    "entra",
		Aliases: map[string]string{},
		Origins: map[string]string{
			"org": "default", "project": "default", "format": "default",
			"auth": "default", "default-type": "default",
		},
		GlobalPath: globalPath,
	}
	if err := cfg.applyFile(globalPath, "global"); err != nil {
		return nil, err
	}
	if local := findLocal(cwd); local != "" {
		cfg.LocalPath = local
		if err := cfg.applyFile(local, "local"); err != nil {
			return nil, err
		}
	}
	if v := os.Getenv("ABWI_ORG"); v != "" {
		cfg.Org, cfg.Origins["org"] = v, "env"
	}
	if v := os.Getenv("ABWI_PROJECT"); v != "" {
		cfg.Project, cfg.Origins["project"] = v, "env"
	}
	return cfg, cfg.Validate()
}

func (c *Config) applyFile(path, origin string) error {
	var fc fileConfig
	if _, err := toml.DecodeFile(path, &fc); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("%s: %w", path, err)
	}
	set := func(dst *string, v, key string) {
		if v != "" {
			*dst, c.Origins[key] = v, origin
		}
	}
	set(&c.Org, fc.Org, "org")
	set(&c.Project, fc.Project, "project")
	set(&c.Format, fc.Format, "format")
	// auth is only honored in the global config (or the --auth flag): a
	// checked-out repository must not be able to switch the auth mode.
	if fc.Auth != "" && origin == "local" {
		fmt.Fprintf(os.Stderr,
			"abwi: warning: ignoring \"auth\" in %s; set it in %s or pass --auth\n", path, c.GlobalPath)
	} else {
		set(&c.Auth, fc.Auth, "auth")
	}
	set(&c.DefaultType, fc.DefaultType, "default-type")
	for k, v := range fc.Aliases {
		c.Aliases[k] = v
	}
	return nil
}

// Validate rejects unknown format/auth values. It runs after every source is
// applied, including flag overrides.
func (c *Config) Validate() error {
	if c.Format != "markdown" && c.Format != "html" {
		return fmt.Errorf("invalid format %q: must be \"markdown\" or \"html\"", c.Format)
	}
	if c.Auth != "entra" && c.Auth != "pat" {
		return fmt.Errorf("invalid auth %q: must be \"entra\" or \"pat\"", c.Auth)
	}
	return nil
}

// findLocal walks up from dir looking for .abwi.toml.
func findLocal(dir string) string {
	for {
		p := filepath.Join(dir, localName)
		if _, err := os.Stat(p); err == nil {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
