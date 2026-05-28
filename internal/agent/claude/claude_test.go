package claude

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/StefanIE1205/lark-agent-bridge/internal/agent"
)

func TestNewDefaultCommand(t *testing.T) {
	a := New("", nil)
	if a.command != "claude" {
		t.Errorf("default command = %q, want claude", a.command)
	}
	if a.Name() != "claude" {
		t.Errorf("Name() = %q, want claude", a.Name())
	}
}

func TestNewCustomCommand(t *testing.T) {
	a := New("/custom/claude", []string{"--verbose"})
	if a.command != "/custom/claude" {
		t.Errorf("command = %q", a.command)
	}
	if len(a.args) != 1 || a.args[0] != "--verbose" {
		t.Errorf("args = %v", a.args)
	}
}

func TestCheckCommandExists(t *testing.T) {
	a := New("cmd.exe", nil)
	if err := a.Check(context.Background()); err != nil {
		t.Errorf("cmd.exe should exist: %v", err)
	}
}

func TestCheckCommandNotFound(t *testing.T) {
	a := New("nonexistent_claude_xyz", nil)
	if err := a.Check(context.Background()); err == nil {
		t.Error("expected error for nonexistent command")
	}
}

func TestStartSession(t *testing.T) {
	a := New("cmd.exe", []string{"/c", "echo hello from claude"})
	sess, err := a.StartSession(context.Background(), agent.SessionOptions{
		SessionKey: "test",
		WorkDir:    ".",
	})
	if err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sess.Send(ctx, ""); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	var output strings.Builder
	done := false
	for event := range sess.Events() {
		switch event.Type {
		case agent.EventTextDelta:
			output.WriteString(event.Text)
		case agent.EventDone:
			done = true
		}
	}

	if !done {
		t.Error("expected done event")
	}
	if !strings.Contains(output.String(), "hello from claude") {
		t.Errorf("expected 'hello from claude' in output, got: %s", output.String())
	}
}
