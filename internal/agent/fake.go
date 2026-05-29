package agent

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type FakeAgent struct {
	mode string
}

func NewFakeAgent(mode string) *FakeAgent {
	return &FakeAgent{mode: mode}
}

func (a *FakeAgent) Name() string {
	return "fake"
}

func (a *FakeAgent) Capabilities() Capability {
	return Capability{
		PersistentSession: false,
		SupportsApproval:  false,
		SupportsStreaming: true,
		Experimental:      false,
	}
}

func (a *FakeAgent) Check(ctx context.Context) error {
	return nil
}

func (a *FakeAgent) StartSession(ctx context.Context, opts SessionOptions) (AgentSession, error) {
	return &FakeSession{
		agentMode: a.mode,
		opts:      opts,
		events:    make(chan AgentEvent, 100),
	}, nil
}

type FakeSession struct {
	agentMode string
	opts      SessionOptions
	events    chan AgentEvent
	alive     bool
	cancel    context.CancelFunc
}

func (s *FakeSession) Send(ctx context.Context, prompt string) error {
	ctx, s.cancel = context.WithCancel(ctx)
	s.alive = true

	switch s.agentMode {
	case "error":
		go s.runError(ctx)
	case "long":
		go s.runLong(ctx, prompt)
	default:
		go s.runNormal(ctx, prompt)
	}
	return nil
}

func (s *FakeSession) Events() <-chan AgentEvent {
	return s.events
}

func (s *FakeSession) Stop(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	s.alive = false
	return nil
}

func (s *FakeSession) Alive() bool {
	return s.alive
}

// WorkDir returns the work directory passed to this session.
func (s *FakeSession) WorkDir() string {
	return s.opts.WorkDir
}

func (s *FakeSession) runNormal(ctx context.Context, prompt string) {
	defer close(s.events)
	defer func() { s.alive = false }()

	// Clean prompt for display
	prompt = strings.TrimSpace(prompt)
	if len(prompt) > 60 {
		prompt = prompt[:57] + "..."
	}

	words := []string{
		fmt.Sprintf("收到任务：%s\n", prompt),
		"正在分析...\n",
		"任务完成。\n",
	}

	for _, w := range words {
		select {
		case <-ctx.Done():
			s.emit(AgentEvent{Type: EventError, Error: "cancelled", CreatedAt: time.Now()})
			return
		case s.events <- AgentEvent{Type: EventTextDelta, Text: w, CreatedAt: time.Now()}:
		}
		time.Sleep(50 * time.Millisecond)
	}

	s.emit(AgentEvent{Type: EventDone, CreatedAt: time.Now()})
}

func (s *FakeSession) runLong(ctx context.Context, prompt string) {
	defer close(s.events)
	defer func() { s.alive = false }()

	s.emit(AgentEvent{
		Type:      EventTextDelta,
		Text:      fmt.Sprintf("开始长任务：%s\n", strings.TrimSpace(prompt)),
		CreatedAt: time.Now(),
	})

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	count := 0

	for {
		select {
		case <-ctx.Done():
			s.emit(AgentEvent{Type: EventError, Error: "cancelled", CreatedAt: time.Now()})
			return
		case <-ticker.C:
			count++
			s.emit(AgentEvent{
				Type:      EventTextDelta,
				Text:      fmt.Sprintf("处理中... (%d)\n", count),
				CreatedAt: time.Now(),
			})
		}
	}
}

func (s *FakeSession) runError(ctx context.Context) {
	defer close(s.events)
	defer func() { s.alive = false }()

	s.emit(AgentEvent{
		Type:      EventTextDelta,
		Text:      "正在处理...",
		CreatedAt: time.Now(),
	})
	time.Sleep(100 * time.Millisecond)

	s.emit(AgentEvent{
		Type:      EventError,
		Error:     "模拟错误：无法完成此任务",
		CreatedAt: time.Now(),
	})
}

func (s *FakeSession) emit(e AgentEvent) {
	select {
	case s.events <- e:
	default:
	}
}
