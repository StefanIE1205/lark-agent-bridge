package antigravity

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/StefanIE1205/lark-agent-bridge/internal/agent"
)

type Adapter struct {
	command      string
	args         []string
	experimental bool
}

func New(command string, args []string) *Adapter {
	if command == "" {
		command = "agy"
	}
	return &Adapter{command: command, args: args, experimental: true}
}

func (a *Adapter) Name() string {
	return "antigravity"
}

func (a *Adapter) IsExperimental() bool {
	return a.experimental
}

func (a *Adapter) Check(ctx context.Context) error {
	// Try common names
	for _, candidate := range []string{a.command, "agy", "antigravity"} {
		if _, err := exec.LookPath(candidate); err == nil {
			a.command = candidate
			return nil
		}
	}
	return fmt.Errorf("antigravity: no CLI found (checked: %s, agy, antigravity). "+
		"Antigravity CLI 当前无法被本服务稳定驱动，请检查 CLI headless/PTY 支持", a.command)
}

func (a *Adapter) StartSession(ctx context.Context, opts agent.SessionOptions) (agent.AgentSession, error) {
	cfg := agent.RunnerConfig{
		Command: a.command,
		Args:    a.args,
		WorkDir: opts.WorkDir,
		Env:     opts.Env,
	}

	runner := agent.NewRunner(cfg)
	return runner, nil
}
