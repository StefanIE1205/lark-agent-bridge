package agent

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunnerSimpleCommand(t *testing.T) {
	runner := NewRunner(RunnerConfig{
		Command: "cmd.exe",
		Args:    []string{"/c", "echo hello world"},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := runner.Send(ctx, ""); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	var output strings.Builder
	done := false
	for event := range runner.Events() {
		switch event.Type {
		case EventTextDelta:
			output.WriteString(event.Text)
		case EventDone:
			done = true
		case EventError:
			t.Logf("error event: %s", event.Error)
		}
	}

	if !done {
		t.Error("expected done event")
	}

	out := output.String()
	// On Windows, cmd.exe /c echo outputs "hello world" with a trailing newline
	if !strings.Contains(out, "hello world") {
		t.Errorf("expected 'hello world' in output, got: %s", out)
	}
}

func TestRunnerCancel(t *testing.T) {
	// Use ping -t to create a long-running process on Windows
	runner := NewRunner(RunnerConfig{
		Command: "ping.exe",
		Args:    []string{"-n", "100", "127.0.0.1"},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := runner.Send(ctx, ""); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if !runner.Alive() {
		t.Error("runner should be alive after Send")
	}

	// Let it produce some output
	time.Sleep(300 * time.Millisecond)

	// Cancel should kill the process
	cancel()

	// Wait for events to close
	gotError := false
	for event := range runner.Events() {
		if event.Type == EventError && strings.Contains(event.Error, "cancelled") {
			gotError = true
		}
	}

	if !gotError {
		t.Error("expected cancelled error event after cancel")
	}
	if runner.Alive() {
		t.Error("runner should not be alive after cancel")
	}
}

func TestRunnerStop(t *testing.T) {
	runner := NewRunner(RunnerConfig{
		Command: "ping.exe",
		Args:    []string{"-n", "100", "127.0.0.1"},
	})

	ctx := context.Background()

	if err := runner.Send(ctx, ""); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if !runner.Alive() {
		t.Error("runner should be alive after Send")
	}

	// Stop should kill
	if err := runner.Stop(ctx); err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	if runner.Alive() {
		t.Error("runner should not be alive after Stop")
	}
}

func TestRunnerDoubleSend(t *testing.T) {
	// Use a long-running command so it stays alive
	runner := NewRunner(RunnerConfig{
		Command: "ping.exe",
		Args:    []string{"-n", "100", "127.0.0.1"},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := runner.Send(ctx, ""); err != nil {
		t.Fatalf("first Send failed: %v", err)
	}

	// Second Send while alive should fail
	time.Sleep(200 * time.Millisecond)
	err := runner.Send(ctx, "")
	if err == nil {
		t.Error("expected error on second Send while alive")
	}
}

func TestRunnerCommandNotFound(t *testing.T) {
	runner := NewRunner(RunnerConfig{
		Command: "nonexistent_command_xyz",
	})

	ctx := context.Background()
	err := runner.Send(ctx, "")
	if err == nil {
		t.Error("expected error for nonexistent command")
	}
}

func TestRunnerStderr(t *testing.T) {
	// Redirect stdout to stderr to test stderr capture
	runner := NewRunner(RunnerConfig{
		Command: "cmd.exe",
		Args:    []string{"/c", "echo error message>&2"},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := runner.Send(ctx, ""); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	var output strings.Builder
	for event := range runner.Events() {
		if event.Type == EventTextDelta {
			output.WriteString(event.Text)
		}
	}

	out := output.String()
	if !strings.Contains(out, "[stderr]") {
		t.Errorf("expected stderr prefix in output, got: %s", out)
	}
	if !strings.Contains(out, "error message") {
		t.Errorf("expected error message in output, got: %s", out)
	}
}
