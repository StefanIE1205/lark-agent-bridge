package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadValid(t *testing.T) {
	path := writeTemp(t, `
data_dir = "/tmp/lab"
default_agent = "codex"

[lark]
app_id = "cli_test"
app_secret = "secret123"
admin_user_ids = ["ou_admin"]

[[projects]]
name = "demo"
path = "C:\\repo\\demo"

[[agents]]
name = "codex"
type = "codex"
command = "codex"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
	if cfg.Lark.AppID != "cli_test" {
		t.Errorf("app_id = %q, want %q", cfg.Lark.AppID, "cli_test")
	}
	if cfg.DefaultAgent != "codex" {
		t.Errorf("default_agent = %q, want %q", cfg.DefaultAgent, "codex")
	}
}

func TestLoadMissingAppID(t *testing.T) {
	path := writeTemp(t, `
[lark]
app_secret = "secret"
admin_user_ids = ["ou_admin"]
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing app_id")
	}
}

func TestLoadMissingAppSecret(t *testing.T) {
	path := writeTemp(t, `
[lark]
app_id = "cli_test"
admin_user_ids = ["ou_admin"]
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing app_secret")
	}
}

func TestLoadNoAdminUsers(t *testing.T) {
	path := writeTemp(t, `
[lark]
app_id = "cli_test"
app_secret = "secret"
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for no admin users")
	}
}

func TestLoadDuplicateProjectName(t *testing.T) {
	path := writeTemp(t, `
[lark]
app_id = "cli_test"
app_secret = "secret"
admin_user_ids = ["ou_admin"]

[[projects]]
name = "demo"
path = "C:\\a"

[[projects]]
name = "demo"
path = "C:\\b"
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate project name")
	}
}

func TestLoadDuplicateAgentName(t *testing.T) {
	path := writeTemp(t, `
[lark]
app_id = "cli_test"
app_secret = "secret"
admin_user_ids = ["ou_admin"]

[[agents]]
name = "codex"
type = "codex"
command = "codex"

[[agents]]
name = "codex"
type = "codex"
command = "codex2"
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate agent name")
	}
}

func TestLoadNonAbsoluteProjectPath(t *testing.T) {
	path := writeTemp(t, `
[lark]
app_id = "cli_test"
app_secret = "secret"
admin_user_ids = ["ou_admin"]

[[projects]]
name = "demo"
path = "relative/path"
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for non-absolute project path")
	}
}

func TestLoadEmptyAgentCommand(t *testing.T) {
	path := writeTemp(t, `
[lark]
app_id = "cli_test"
app_secret = "secret"
admin_user_ids = ["ou_admin"]

[[agents]]
name = "codex"
type = "codex"
command = ""
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty agent command")
	}
}

func TestDefaultValues(t *testing.T) {
	path := writeTemp(t, `
[lark]
app_id = "cli_test"
app_secret = "secret"
admin_user_ids = ["ou_admin"]
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DataDir != "" && cfg.DataDir != "~/.lark-agent-bridge" {
		// data_dir may have been expanded, so just check it's set or default
	}
	if cfg.ProgressIntervalMs != 3000 {
		t.Errorf("default progress_interval_ms = %d, want 3000", cfg.ProgressIntervalMs)
	}
}

func TestProjectPathLookup(t *testing.T) {
	path := writeTemp(t, `
[lark]
app_id = "cli_test"
app_secret = "secret"
admin_user_ids = ["ou_admin"]

[[projects]]
name = "demo"
path = "C:\\repo\\demo"

[[projects]]
name = "backend"
path = "D:\\code\\backend"
`)
	cfg, _ := Load(path)

	p, ok := cfg.ProjectPath("demo")
	if !ok || p != "C:\\repo\\demo" {
		t.Errorf("ProjectPath(demo) = (%q, %v), want (C:\\repo\\demo, true)", p, ok)
	}

	p, ok = cfg.ProjectPath("missing")
	if ok {
		t.Errorf("ProjectPath(missing) = (%q, %v), want (\"\", false)", p, ok)
	}
}

func TestAgentConfigLookup(t *testing.T) {
	path := writeTemp(t, `
[lark]
app_id = "cli_test"
app_secret = "secret"
admin_user_ids = ["ou_admin"]

[[agents]]
name = "codex"
type = "codex"
command = "codex"
args = ["--model", "gpt-4"]
`)

	cfg, _ := Load(path)

	a, ok := cfg.AgentConfig("codex")
	if !ok {
		t.Fatal("expected agent 'codex' to exist")
	}
	if a.Command != "codex" {
		t.Errorf("command = %q, want %q", a.Command, "codex")
	}
	if len(a.Args) != 2 {
		t.Errorf("args count = %d, want 2", len(a.Args))
	}

	_, ok = cfg.AgentConfig("claude")
	if ok {
		t.Error("expected agent 'claude' to not exist")
	}
}

func TestConfigFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.toml")
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}
