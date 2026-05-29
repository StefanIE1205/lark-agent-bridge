package conversation

import (
	"testing"
)

func TestMemoryGetEmpty(t *testing.T) {
	m := NewMemory()
	ctx := m.Get("oc_chat", "ot_thread")

	if ctx.ChatID != "oc_chat" {
		t.Errorf("ChatID = %q, want oc_chat", ctx.ChatID)
	}
	if ctx.ThreadID != "ot_thread" {
		t.Errorf("ThreadID = %q, want ot_thread", ctx.ThreadID)
	}
	if ctx.LastProject != "" {
		t.Errorf("LastProject should be empty, got %q", ctx.LastProject)
	}
}

func TestMemorySetProject(t *testing.T) {
	m := NewMemory()
	m.SetProject("oc_chat", "ot_thread", "backend")

	ctx := m.Get("oc_chat", "ot_thread")
	if ctx.LastProject != "backend" {
		t.Errorf("LastProject = %q, want backend", ctx.LastProject)
	}
}

func TestMemorySetAgent(t *testing.T) {
	m := NewMemory()
	m.SetAgent("oc_chat", "ot_thread", "claude")

	ctx := m.Get("oc_chat", "ot_thread")
	if ctx.LastAgent != "claude" {
		t.Errorf("LastAgent = %q, want claude", ctx.LastAgent)
	}
}

func TestMemorySetTask(t *testing.T) {
	m := NewMemory()
	m.SetTask("oc_chat", "ot_thread", "fix bug", "running")

	ctx := m.Get("oc_chat", "ot_thread")
	if ctx.LastTask != "fix bug" {
		t.Errorf("LastTask = %q, want fix bug", ctx.LastTask)
	}
	if ctx.LastTaskStatus != "running" {
		t.Errorf("LastTaskStatus = %q, want running", ctx.LastTaskStatus)
	}
}

func TestMemorySetTaskStatus(t *testing.T) {
	m := NewMemory()
	m.SetTask("oc_chat", "ot_thread", "fix bug", "running")
	m.SetTaskStatus("oc_chat", "ot_thread", "done")

	ctx := m.Get("oc_chat", "ot_thread")
	if ctx.LastTaskStatus != "done" {
		t.Errorf("LastTaskStatus = %q, want done", ctx.LastTaskStatus)
	}
	if ctx.LastTask != "fix bug" {
		t.Errorf("LastTask should still be 'fix bug', got %q", ctx.LastTask)
	}
}

func TestMemoryDifferentThreads(t *testing.T) {
	m := NewMemory()
	m.SetProject("oc_chat", "ot_thread1", "backend")
	m.SetProject("oc_chat", "ot_thread2", "frontend")

	ctx1 := m.Get("oc_chat", "ot_thread1")
	ctx2 := m.Get("oc_chat", "ot_thread2")

	if ctx1.LastProject != "backend" {
		t.Errorf("thread1 project = %q, want backend", ctx1.LastProject)
	}
	if ctx2.LastProject != "frontend" {
		t.Errorf("thread2 project = %q, want frontend", ctx2.LastProject)
	}
}

func TestMemoryNoThread(t *testing.T) {
	m := NewMemory()
	m.SetProject("oc_chat", "", "backend")

	ctx := m.Get("oc_chat", "")
	if ctx.LastProject != "backend" {
		t.Errorf("LastProject = %q, want backend", ctx.LastProject)
	}
}
