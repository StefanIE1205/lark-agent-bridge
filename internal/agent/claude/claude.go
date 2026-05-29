package claude

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/StefanIE1205/lark-agent-bridge/internal/agent"
)

type Adapter struct {
	command string
	args    []string
}

func New(command string, args []string) *Adapter {
	if command == "" {
		command = "claude"
	}
	return &Adapter{command: command, args: args}
}

func (a *Adapter) Name() string {
	return "claude"
}

func (a *Adapter) Capabilities() agent.Capability {
	return agent.Capability{
		PersistentSession: false,
		SupportsApproval:  false,
		SupportsStreaming: true,
		Experimental:      false,
	}
}

func (a *Adapter) Check(ctx context.Context) error {
	_, err := exec.LookPath(a.command)
	if err != nil {
		return fmt.Errorf("claude: command %q not found in PATH: %w", a.command, err)
	}
	return nil
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
