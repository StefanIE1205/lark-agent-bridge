package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DataDir              string   `toml:"data_dir"`
	DefaultAgent         string   `toml:"default_agent"`
	DefaultProject       string   `toml:"default_project"`
	ProgressIntervalMs   int      `toml:"progress_interval_ms"`
	Lark                 Lark     `toml:"lark"`
	Projects             []Project `toml:"projects"`
	Agents               []Agent  `toml:"agents"`
	Security             Security `toml:"security"`
}

type Lark struct {
	AppID                 string   `toml:"app_id"`
	AppSecret             string   `toml:"app_secret"`
	AdminUserIDs          []string `toml:"admin_user_ids"`
	AllowedChatIDs        []string `toml:"allowed_chat_ids"`
	RequireMentionInGroup bool     `toml:"require_mention_in_group"`
}

type Project struct {
	Name string `toml:"name"`
	Path string `toml:"path"`
}

type Agent struct {
	Name         string   `toml:"name"`
	Type         string   `toml:"type"`
	Command      string   `toml:"command"`
	Args         []string `toml:"args"`
	Experimental bool     `toml:"experimental"`
}

type Security struct {
	AllowShell              bool `toml:"allow_shell"`
	RequireConfirmForWrite  bool `toml:"require_confirm_for_write"`
	RequireConfirmForInstall bool `toml:"require_confirm_for_install"`
	RequireConfirmForGitPush bool `toml:"require_confirm_for_git_push"`
	RedactSecrets           bool `toml:"redact_secrets"`
}

func Load(path string) (*Config, error) {
	filePath := resolvePath(path)
	if filePath == "" {
		return nil, fmt.Errorf("config: no config file found (looked in: --config flag, ./config.toml, ~/.lark-agent-bridge/config.toml)")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", filePath, err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", filePath, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config: validate %s: %w", filePath, err)
	}

	cfg.DataDir = expandHome(cfg.DataDir)

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.DataDir == "" {
		c.DataDir = "~/.lark-agent-bridge"
	}

	if c.ProgressIntervalMs <= 0 {
		c.ProgressIntervalMs = 3000
	}

	if c.Lark.AppID == "" {
		return fmt.Errorf("lark.app_id is required")
	}
	if c.Lark.AppSecret == "" {
		return fmt.Errorf("lark.app_secret is required")
	}
	if len(c.Lark.AdminUserIDs) == 0 {
		return fmt.Errorf("at least one lark.admin_user_ids entry is required")
	}

	projectNames := make(map[string]bool)
	for i, p := range c.Projects {
		if p.Name == "" {
			return fmt.Errorf("projects[%d]: name is required", i)
		}
		if projectNames[p.Name] {
			return fmt.Errorf("projects[%d]: duplicate project name %q", i, p.Name)
		}
		projectNames[p.Name] = true
		if p.Path == "" {
			return fmt.Errorf("projects[%d]: path is required", i)
		}
		if !filepath.IsAbs(p.Path) {
			return fmt.Errorf("projects[%d]: path %q must be absolute", i, p.Path)
		}
	}

	agentNames := make(map[string]bool)
	for i, a := range c.Agents {
		if a.Name == "" {
			return fmt.Errorf("agents[%d]: name is required", i)
		}
		if agentNames[a.Name] {
			return fmt.Errorf("agents[%d]: duplicate agent name %q", i, a.Name)
		}
		agentNames[a.Name] = true
		if a.Command == "" {
			return fmt.Errorf("agents[%d]: command is required", i)
		}
	}

	return nil
}

func (c *Config) ProjectPath(name string) (string, bool) {
	for _, p := range c.Projects {
		if p.Name == name {
			return p.Path, true
		}
	}
	return "", false
}

func (c *Config) AgentConfig(name string) (Agent, bool) {
	for _, a := range c.Agents {
		if a.Name == name {
			return a, true
		}
	}
	return Agent{}, false
}

func resolvePath(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if _, err := os.Stat("config.toml"); err == nil {
		return "config.toml"
	}
	home := expandHome("~/.lark-agent-bridge/config.toml")
	if _, err := os.Stat(home); err == nil {
		return home
	}
	return ""
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		if path == "~" {
			return home
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
