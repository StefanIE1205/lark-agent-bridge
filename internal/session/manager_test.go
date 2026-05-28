package session

import (
	"testing"
)

func TestSessionKey(t *testing.T) {
	key := DeriveKey("oc_chat", "ot_thread", "demo", "codex")
	expected := "lark:oc_chat:ot_thread:demo:codex"
	if key.String() != expected {
		t.Errorf("key = %q, want %q", key.String(), expected)
	}
}

func TestSessionKeyThreadFallback(t *testing.T) {
	key := DeriveKey("oc_chat", "", "demo", "claude")
	if key.ThreadID != "oc_chat" {
		t.Errorf("ThreadID fallback = %q, want oc_chat", key.ThreadID)
	}
}

func TestGetOrCreate(t *testing.T) {
	m := NewManager()

	s1, err := m.GetOrCreate("oc_chat", "ot_thread", "demo", "codex", "C:\\demo")
	if err != nil {
		t.Fatal(err)
	}
	if s1.Status != StatusIdle {
		t.Errorf("expected idle, got %s", s1.Status)
	}

	// Same params should return the same session
	s2, err := m.GetOrCreate("oc_chat", "ot_thread", "demo", "codex", "C:\\demo")
	if err != nil {
		t.Fatal(err)
	}
	if s1.Key != s2.Key {
		t.Errorf("expected same session key: %s vs %s", s1.Key, s2.Key)
	}
}

func TestGetOrCreateDifferentParams(t *testing.T) {
	m := NewManager()

	s1, _ := m.GetOrCreate("oc_chat", "ot_thread", "demo", "codex", "C:\\demo")
	s2, _ := m.GetOrCreate("oc_chat", "ot_thread", "demo", "claude", "C:\\demo")

	if s1.Key == s2.Key {
		t.Errorf("different agents should create different sessions")
	}
}

func TestStartTask(t *testing.T) {
	m := NewManager()
	s, _ := m.GetOrCreate("oc_chat", "ot_thread", "demo", "codex", "C:\\demo")

	if err := m.StartTask(s.Key); err != nil {
		t.Fatalf("StartTask failed: %v", err)
	}

	sess := m.Get(s.Key)
	if sess.Status != StatusRunning {
		t.Errorf("expected running, got %s", sess.Status)
	}
}

func TestStartTaskAlreadyRunning(t *testing.T) {
	m := NewManager()
	s, _ := m.GetOrCreate("oc_chat", "ot_thread", "demo", "codex", "C:\\demo")

	m.StartTask(s.Key)
	err := m.StartTask(s.Key)
	if err == nil {
		t.Fatal("expected error when starting already-running session")
	}
}

func TestStop(t *testing.T) {
	m := NewManager()
	s, _ := m.GetOrCreate("oc_chat", "ot_thread", "demo", "codex", "C:\\demo")

	m.StartTask(s.Key)
	if err := m.Stop(s.Key); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	sess := m.Get(s.Key)
	if sess.Status != StatusIdle {
		t.Errorf("expected idle after stop, got %s", sess.Status)
	}
}

func TestStopNotRunning(t *testing.T) {
	m := NewManager()
	s, _ := m.GetOrCreate("oc_chat", "ot_thread", "demo", "codex", "C:\\demo")

	err := m.Stop(s.Key)
	if err == nil {
		t.Fatal("expected error when stopping idle session")
	}
}

func TestTransition(t *testing.T) {
	m := NewManager()
	s, _ := m.GetOrCreate("oc_chat", "ot_thread", "demo", "codex", "C:\\demo")

	if err := m.Transition(s.Key, StatusStarting); err != nil {
		t.Fatalf("Transition to starting failed: %v", err)
	}
	if err := m.Transition(s.Key, StatusRunning); err != nil {
		t.Fatalf("Transition to running failed: %v", err)
	}
	if err := m.Transition(s.Key, StatusFailed); err != nil {
		t.Fatalf("Transition to failed failed: %v", err)
	}
	if err := m.Transition(s.Key, StatusIdle); err != nil {
		t.Fatalf("Transition to idle failed: %v", err)
	}
}

func TestTransitionInvalid(t *testing.T) {
	m := NewManager()
	s, _ := m.GetOrCreate("oc_chat", "ot_thread", "demo", "codex", "C:\\demo")

	// Cannot go directly from idle to stopping
	err := m.Transition(s.Key, StatusStopping)
	if err == nil {
		t.Fatal("expected error for invalid transition idle -> stopping")
	}
}

func TestListActive(t *testing.T) {
	m := NewManager()
	s1, _ := m.GetOrCreate("oc_1", "ot_1", "demo", "codex", "C:\\demo")
	s2, _ := m.GetOrCreate("oc_2", "ot_2", "backend", "claude", "C:\\backend")

	m.StartTask(s1.Key)

	active := m.ListActive()
	if len(active) != 2 {
		t.Errorf("expected 2 active sessions, got %d", len(active))
	}

	// Close s2
	m.Transition(s2.Key, StatusClosed)
	active = m.ListActive()
	if len(active) != 1 {
		t.Errorf("expected 1 active session after close, got %d", len(active))
	}
}

func TestSetCancelFunc(t *testing.T) {
	m := NewManager()
	s, _ := m.GetOrCreate("oc", "ot", "demo", "codex", "C:\\demo")

	cancelled := false
	m.SetCancelFunc(s.Key, func() {
		cancelled = true
	})

	m.StartTask(s.Key)
	m.Stop(s.Key)

	if !cancelled {
		t.Error("cancel function was not called on stop")
	}
}

func TestSetError(t *testing.T) {
	m := NewManager()
	s, _ := m.GetOrCreate("oc", "ot", "demo", "codex", "C:\\demo")

	m.SetError(s.Key, "something went wrong")

	sess := m.Get(s.Key)
	if sess.LastError != "something went wrong" {
		t.Errorf("LastError = %q, want %q", sess.LastError, "something went wrong")
	}
}

func TestIsProjectRunning(t *testing.T) {
	m := NewManager()
	s1, _ := m.GetOrCreate("oc_1", "ot_1", "demo", "codex", "C:\\demo")
	_, _ = m.GetOrCreate("oc_2", "ot_2", "backend", "claude", "C:\\backend")

	m.StartTask(s1.Key)

	if !m.IsProjectRunning("demo") {
		t.Error("demo project should be running")
	}
	if m.IsProjectRunning("backend") {
		t.Error("backend project should not be running")
	}
}

func TestCanTransition(t *testing.T) {
	if !CanTransition(StatusIdle, StatusStarting) {
		t.Error("idle -> starting should be valid")
	}
	if !CanTransition(StatusRunning, StatusFailed) {
		t.Error("running -> failed should be valid")
	}
	if CanTransition(StatusIdle, StatusStopping) {
		t.Error("idle -> stopping should be invalid")
	}
	if CanTransition(StatusClosed, StatusRunning) {
		t.Error("closed -> running should be invalid")
	}
}
