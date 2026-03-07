package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ActivityRule struct {
	Match    []string `yaml:"match"`
	Activity string   `yaml:"activity"`
}

type Config struct {
	Endpoint      string         `yaml:"endpoint"`
	Token         string         `yaml:"token"`
	Template      string         `yaml:"template"`
	Interval      string         `yaml:"interval,omitempty"`
	ActivityRules []ActivityRule `yaml:"activity_rules,omitempty"`
	Ignore        []string       `yaml:"ignore,omitempty"`
	Telemetry     *bool          `yaml:"telemetry,omitempty"`
}

// TelemetryEnabled returns true unless explicitly disabled.
func (c Config) TelemetryEnabled() bool {
	return c.Telemetry == nil || *c.Telemetry
}

func DefaultConfig() Config {
	return Config{
		Endpoint: "https://now.ctx.st",
		Template: "{activity}",
		Interval: "30s",
		ActivityRules: []ActivityRule{
			{Match: []string{"Visual Studio Code", "Code", "Cursor", "Windsurf", "Zed"}, Activity: "Coding"},
			{Match: []string{"Terminal", "iTerm2", "Warp", "Alacritty", "kitty"}, Activity: "In terminal"},
			{Match: []string{"Google Chrome", "Safari", "Arc", "Firefox", "Brave Browser", "Microsoft Edge"}, Activity: "Browsing"},
			{Match: []string{"Figma", "Sketch"}, Activity: "Designing"},
			{Match: []string{"Slack", "Discord", "Telegram", "WeChat", "Messages"}, Activity: "Chatting"},
			{Match: []string{"Notion", "Obsidian", "Bear", "Notes"}, Activity: "Writing"},
		},
		Ignore: []string{"1Password", "System Preferences", "System Settings"},
	}
}

// Dir returns the config directory path.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}

	// Respect XDG_CONFIG_HOME if set
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "nownow"), nil
	}
	return filepath.Join(home, ".config", "nownow"), nil
}

// Path returns the full path to config.yml.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yml"), nil
}

// Load reads config from disk. Returns default config if file doesn't exist.
func Load() (Config, error) {
	cfg := DefaultConfig()

	p, err := Path()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	// Ensure defaults for empty fields
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://now.ctx.st"
	}
	if cfg.Template == "" {
		cfg.Template = "{activity}"
	}
	if cfg.Interval == "" {
		cfg.Interval = "30s"
	}

	// Migrate legacy templates: strip removed {project}/{branch} placeholders
	cfg.Template = migrateTemplate(cfg.Template)

	return cfg, nil
}

// Save writes config to disk, creating the directory if needed.
func Save(cfg Config) error {
	p, err := Path()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(p, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// HasToken returns true if a token is configured.
func (c Config) HasToken() bool {
	return c.Token != ""
}

// IsIgnored returns true if the app name should be ignored.
func (c Config) IsIgnored(app string) bool {
	for _, name := range c.Ignore {
		if name == app {
			return true
		}
	}
	return false
}

// ActivityFor returns the activity label for a given app name via exact case-insensitive match.
func (c Config) ActivityFor(app string) string {
	for _, rule := range c.ActivityRules {
		for _, m := range rule.Match {
			if strings.EqualFold(app, m) {
				return rule.Activity
			}
		}
	}
	return ""
}

// ResolveActivity builds the full activity string with watching/music context.
// Priority: watching > matched activity > "Using {app}", with music appended when not watching.
// Returns "" if no meaningful activity can be determined.
func (c Config) ResolveActivity(app, watching, music string) string {
	activity := c.ActivityFor(app)

	if watching != "" {
		activity = "Watching: " + watching
	} else if activity == "" && app != "" {
		activity = "Using " + app
	}

	if music != "" && watching == "" && activity != "" {
		activity = activity + " · Listening to " + music
	}

	return activity
}

// migrateTemplate strips removed {project}/{branch} placeholders and legacy emoji references from templates.
func migrateTemplate(tmpl string) string {
	// Migrate legacy emoji placeholders to activity
	tmpl = strings.ReplaceAll(tmpl, "{emoji} {app}", "{activity}")
	tmpl = strings.ReplaceAll(tmpl, "{emoji}", "{activity}")

	tmpl = strings.ReplaceAll(tmpl, "{project}", "")
	tmpl = strings.ReplaceAll(tmpl, "{branch}", "")

	// Clean up artifacts: empty parens/brackets, collapse spaces first, then separators
	tmpl = strings.ReplaceAll(tmpl, "()", "")
	tmpl = strings.ReplaceAll(tmpl, "[]", "")
	for strings.Contains(tmpl, "  ") {
		tmpl = strings.ReplaceAll(tmpl, "  ", " ")
	}
	for strings.Contains(tmpl, "· ·") {
		tmpl = strings.ReplaceAll(tmpl, "· ·", "·")
	}
	// Re-collapse spaces after middot cleanup
	for strings.Contains(tmpl, "  ") {
		tmpl = strings.ReplaceAll(tmpl, "  ", " ")
	}
	tmpl = strings.TrimRight(tmpl, " ·")
	tmpl = strings.TrimLeft(tmpl, " ·")
	return strings.TrimSpace(tmpl)
}
