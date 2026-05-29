package conversation

import (
	"testing"
	"time"
)

func TestClarificationSaveAndGet(t *testing.T) {
	m := NewClarificationManager(5 * time.Minute)

	m.Save("oc_chat", "ot_thread", ClarifyProject, "选哪个项目？", []string{"backend", "frontend"}, "fix bug")

	c := m.Get("oc_chat", "ot_thread")
	if c == nil {
		t.Fatal("expected clarification to exist")
	}
	if c.Type != ClarifyProject {
		t.Errorf("type = %q, want project", c.Type)
	}
	if c.Question != "选哪个项目？" {
		t.Errorf("question = %q, want '选哪个项目？'", c.Question)
	}
	if c.Task != "fix bug" {
		t.Errorf("task = %q, want 'fix bug'", c.Task)
	}
}

func TestClarificationClear(t *testing.T) {
	m := NewClarificationManager(5 * time.Minute)

	m.Save("oc_chat", "", ClarifyProject, "question", nil, "task")
	m.Clear("oc_chat", "")

	if m.HasPending("oc_chat", "") {
		t.Error("should not have pending after clear")
	}
}

func TestClarificationExpired(t *testing.T) {
	m := NewClarificationManager(1 * time.Millisecond)

	m.Save("oc_chat", "", ClarifyProject, "question", nil, "task")
	time.Sleep(5 * time.Millisecond)

	c := m.Get("oc_chat", "")
	if c != nil {
		t.Error("should return nil for expired clarification")
	}
}

func TestClarificationHasPending(t *testing.T) {
	m := NewClarificationManager(5 * time.Minute)

	if m.HasPending("oc_chat", "") {
		t.Error("should not have pending initially")
	}

	m.Save("oc_chat", "", ClarifyProject, "question", nil, "task")

	if !m.HasPending("oc_chat", "") {
		t.Error("should have pending after save")
	}
}

func TestClarificationDifferentThreads(t *testing.T) {
	m := NewClarificationManager(5 * time.Minute)

	m.Save("oc_chat", "ot_thread1", ClarifyProject, "q1", nil, "task1")
	m.Save("oc_chat", "ot_thread2", ClarifyAgent, "q2", nil, "task2")

	c1 := m.Get("oc_chat", "ot_thread1")
	c2 := m.Get("oc_chat", "ot_thread2")

	if c1.Type != ClarifyProject {
		t.Errorf("thread1 type = %q, want project", c1.Type)
	}
	if c2.Type != ClarifyAgent {
		t.Errorf("thread2 type = %q, want agent", c2.Type)
	}
}
