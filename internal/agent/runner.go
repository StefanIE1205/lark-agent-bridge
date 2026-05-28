package agent

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"
)

type RunnerConfig struct {
	Command string
	Args    []string
	WorkDir string
	Env     []string
}

type Runner struct {
	config RunnerConfig
	events chan AgentEvent
	cancel context.CancelFunc
	alive  bool
}

func NewRunner(cfg RunnerConfig) *Runner {
	return &Runner{
		config: cfg,
		events: make(chan AgentEvent, 100),
	}
}

func (r *Runner) Send(ctx context.Context, prompt string) error {
	if r.alive {
		return fmt.Errorf("runner: session already active")
	}

	ctx, r.cancel = context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, r.config.Command, r.config.Args...)
	cmd.Dir = r.config.WorkDir
	cmd.Env = r.config.Env

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("runner: stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("runner: stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("runner: stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		close(r.events)
		return fmt.Errorf("runner: start %s: %w", r.config.Command, err)
	}

	r.alive = true

	go func() {
		defer close(r.events)
		defer func() { r.alive = false }()

		// Write prompt to stdin
		if _, err := io.WriteString(stdin, prompt); err == nil {
			stdin.Close()
		}

		// Read stdout and stderr in parallel
		done := make(chan struct{}, 2)

		go func() {
			r.scanLines(stdout, false)
			done <- struct{}{}
		}()
		go func() {
			r.scanLines(stderr, true)
			done <- struct{}{}
		}()

		// Wait for both readers
		<-done
		<-done

		err := cmd.Wait()

		if ctx.Err() != nil {
			r.emit(AgentEvent{Type: EventError, Error: "cancelled", CreatedAt: time.Now()})
			return
		}

		if err != nil {
			r.emit(AgentEvent{
				Type:      EventError,
				Error:     fmt.Sprintf("process exited: %v", err),
				CreatedAt: time.Now(),
			})
			return
		}

		r.emit(AgentEvent{Type: EventDone, CreatedAt: time.Now()})
	}()

	return nil
}

func (r *Runner) Events() <-chan AgentEvent {
	return r.events
}

func (r *Runner) Stop(ctx context.Context) error {
	if r.cancel != nil {
		r.cancel()
	}
	r.alive = false
	return nil
}

func (r *Runner) Alive() bool {
	return r.alive
}

func (r *Runner) scanLines(reader io.Reader, isStderr bool) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		text := scanner.Text()
		if isStderr {
			text = "[stderr] " + text
		}
		r.emit(AgentEvent{
			Type:      EventTextDelta,
			Text:      text + "\n",
			CreatedAt: time.Now(),
		})
	}
}

func (r *Runner) emit(e AgentEvent) {
	select {
	case r.events <- e:
	default:
	}
}
