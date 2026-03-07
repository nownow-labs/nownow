package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Endpoint != "https://now.ctx.st" {
		t.Errorf("expected endpoint https://now.ctx.st, got %s", cfg.Endpoint)
	}
	if cfg.Interval != "30s" {
		t.Errorf("expected interval 30s, got %s", cfg.Interval)
	}
	if cfg.Template == "" {
		t.Error("expected non-empty template")
	}
	if len(cfg.ActivityRules) == 0 {
		t.Error("expected default activity rules")
	}
}

func TestHasToken(t *testing.T) {
	cfg := Config{}
	if cfg.HasToken() {
		t.Error("empty config should not have token")
	}
	cfg.Token = "now_abc"
	if !cfg.HasToken() {
		t.Error("config with token should have token")
	}
}

func TestIsIgnored(t *testing.T) {
	cfg := Config{Ignore: []string{"1Password", "System Preferences"}}

	if !cfg.IsIgnored("1Password") {
		t.Error("1Password should be ignored")
	}
	if cfg.IsIgnored("VS Code") {
		t.Error("VS Code should not be ignored")
	}
}

func TestActivityFor(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		app  string
		want string
	}{
		{"Visual Studio Code", "Coding"},
		{"Code", "Coding"},
		{"iTerm2", "In terminal"},
		{"Google Chrome", "Browsing"},
		{"Unknown App", ""},
		{"Codeium", ""},
	}

	for _, tt := range tests {
		got := cfg.ActivityFor(tt.app)
		if got != tt.want {
			t.Errorf("ActivityFor(%q) = %q, want %q", tt.app, got, tt.want)
		}
	}
}

func TestActivityForCaseInsensitive(t *testing.T) {
	cfg := DefaultConfig()

	got := cfg.ActivityFor("code")
	if got != "Coding" {
		t.Errorf("ActivityFor(%q) = %q, want %q", "code", got, "Coding")
	}

	got = cfg.ActivityFor("SAFARI")
	if got != "Browsing" {
		t.Errorf("ActivityFor(%q) = %q, want %q", "SAFARI", got, "Browsing")
	}
}

func TestResolveActivity(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		name     string
		app      string
		watching string
		music    string
		want     string
	}{
		{"watching overrides", "Safari", "Breaking Bad", "", "Watching: Breaking Bad"},
		{"activity matched", "Code", "", "", "Coding"},
		{"activity with music", "Code", "", "Daft Punk - Get Lucky", "Coding · Listening to Daft Punk - Get Lucky"},
		{"no match fallback", "SomeApp", "", "", "Using SomeApp"},
		{"no match with music", "SomeApp", "", "Queen - Radio Ga Ga", "Using SomeApp · Listening to Queen - Radio Ga Ga"},
		{"watching ignores music", "Safari", "Stranger Things", "Daft Punk - Get Lucky", "Watching: Stranger Things"},
		{"empty app", "", "", "", ""},
		{"empty app with music", "", "", "Daft Punk - Get Lucky", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.ResolveActivity(tt.app, tt.watching, tt.music)
			if got != tt.want {
				t.Errorf("ResolveActivity(%q, %q, %q) = %q, want %q", tt.app, tt.watching, tt.music, got, tt.want)
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Use a temp directory
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg := DefaultConfig()
	cfg.Token = "now_test_token_123"

	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify file exists
	p := filepath.Join(tmpDir, "nownow", "config.yml")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	// Verify permissions
	info, _ := os.Stat(p)
	if info.Mode().Perm() != 0600 {
		t.Errorf("config file permissions: got %o, want 600", info.Mode().Perm())
	}

	// Load it back
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Token != "now_test_token_123" {
		t.Errorf("loaded token = %q, want %q", loaded.Token, "now_test_token_123")
	}
	if loaded.Endpoint != "https://now.ctx.st" {
		t.Errorf("loaded endpoint = %q, want https://now.ctx.st", loaded.Endpoint)
	}
}

func TestLoadMissing(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load missing: %v", err)
	}
	// Should return defaults
	if cfg.Endpoint != "https://now.ctx.st" {
		t.Errorf("expected default endpoint, got %s", cfg.Endpoint)
	}
}

func TestLoadPartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	dir := filepath.Join(tmpDir, "nownow")
	os.MkdirAll(dir, 0700)
	// Write a partial config — only token, no endpoint
	os.WriteFile(filepath.Join(dir, "config.yml"), []byte("token: now_partial\n"), 0600)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Token != "now_partial" {
		t.Errorf("token = %q, want now_partial", cfg.Token)
	}
	// Defaults should be filled in
	if cfg.Endpoint != "https://now.ctx.st" {
		t.Errorf("endpoint not defaulted: %s", cfg.Endpoint)
	}
	if cfg.Interval != "30s" {
		t.Errorf("interval not defaulted: %s", cfg.Interval)
	}
}

func TestTelemetryEnabled(t *testing.T) {
	// nil (default) = enabled
	cfg := Config{}
	if !cfg.TelemetryEnabled() {
		t.Error("nil Telemetry should default to enabled")
	}

	// explicitly true
	b := true
	cfg.Telemetry = &b
	if !cfg.TelemetryEnabled() {
		t.Error("Telemetry=true should be enabled")
	}

	// explicitly false
	b2 := false
	cfg.Telemetry = &b2
	if cfg.TelemetryEnabled() {
		t.Error("Telemetry=false should be disabled")
	}
}

func TestMigrateTemplate(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"legacy full", "{emoji} {app} · {project} ({branch})", "{activity}"},
		{"project only", "{app} · {project}", "{app}"},
		{"branch only", "{app} ({branch})", "{app}"},
		{"legacy emoji app", "{emoji} {app}", "{activity}"},
		{"legacy with music", "{app} · {project} · {music}", "{app} · {music}"},
		{"emoji only", "{emoji}", "{activity}"},
		{"no legacy placeholders", "{activity}", "{activity}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := migrateTemplate(tt.in)
			if got != tt.want {
				t.Errorf("migrateTemplate(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestLoadLegacyTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	dir := filepath.Join(tmpDir, "nownow")
	os.MkdirAll(dir, 0700)
	os.WriteFile(filepath.Join(dir, "config.yml"), []byte("token: now_test\ntemplate: \"{emoji} {app} · {project} ({branch})\"\n"), 0600)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Template != "{activity}" {
		t.Errorf("legacy template not migrated: got %q, want %q", cfg.Template, "{activity}")
	}
}
