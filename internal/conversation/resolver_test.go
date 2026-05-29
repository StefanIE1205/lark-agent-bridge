package conversation

import (
	"testing"
)

func TestResolveProjectFromHint(t *testing.T) {
	r := NewResolver([]string{"backend", "frontend"}, nil)
	project, needAsk, _ := r.ResolveProject("backend", "", "")

	if project != "backend" {
		t.Errorf("project = %q, want backend", project)
	}
	if needAsk {
		t.Error("should not need ask when hint matches")
	}
}

func TestResolveProjectFromThreadMemory(t *testing.T) {
	r := NewResolver([]string{"backend", "frontend"}, nil)
	project, needAsk, _ := r.ResolveProject("", "frontend", "")

	if project != "frontend" {
		t.Errorf("project = %q, want frontend", project)
	}
	if needAsk {
		t.Error("should not need ask when thread memory exists")
	}
}

func TestResolveProjectFromChatMemory(t *testing.T) {
	r := NewResolver([]string{"backend", "frontend"}, nil)
	project, needAsk, _ := r.ResolveProject("", "", "backend")

	if project != "backend" {
		t.Errorf("project = %q, want backend", project)
	}
	if needAsk {
		t.Error("should not need ask when chat memory exists")
	}
}

func TestResolveProjectSingleAutoSelect(t *testing.T) {
	r := NewResolver([]string{"only-project"}, nil)
	project, needAsk, _ := r.ResolveProject("", "", "")

	if project != "only-project" {
		t.Errorf("project = %q, want only-project", project)
	}
	if needAsk {
		t.Error("should not need ask with single project")
	}
}

func TestResolveProjectMultipleNeedAsk(t *testing.T) {
	r := NewResolver([]string{"backend", "frontend", "mobile"}, nil)
	_, needAsk, question := r.ResolveProject("", "", "")

	if !needAsk {
		t.Error("should need ask with multiple projects and no context")
	}
	if question == "" {
		t.Error("question should not be empty")
	}
}

func TestResolveProjectNoneConfigured(t *testing.T) {
	r := NewResolver(nil, nil)
	_, _, err := r.ResolveProject("", "", "")

	if err == "" {
		t.Error("should return error when no projects configured")
	}
}

func TestResolveAgentFromHint(t *testing.T) {
	r := NewResolver(nil, []string{"claude", "codex"})
	agent := r.ResolveAgent("claude", "", "", "", "codex")

	if agent != "claude" {
		t.Errorf("agent = %q, want claude", agent)
	}
}

func TestResolveAgentFromThreadMemory(t *testing.T) {
	r := NewResolver(nil, []string{"claude", "codex"})
	agent := r.ResolveAgent("", "codex", "", "", "claude")

	if agent != "codex" {
		t.Errorf("agent = %q, want codex", agent)
	}
}

func TestResolveAgentFromChatMemory(t *testing.T) {
	r := NewResolver(nil, []string{"claude", "codex"})
	agent := r.ResolveAgent("", "", "claude", "", "codex")

	if agent != "claude" {
		t.Errorf("agent = %q, want claude", agent)
	}
}

func TestResolveAgentFromTaskType(t *testing.T) {
	r := NewResolver(nil, []string{"claude", "codex"})
	agent := r.ResolveAgent("", "", "", "review", "codex")

	if agent != "claude" {
		t.Errorf("agent = %q, want claude for review task", agent)
	}
}

func TestResolveAgentDefault(t *testing.T) {
	r := NewResolver(nil, []string{"claude", "codex"})
	agent := r.ResolveAgent("", "", "", "", "codex")

	if agent != "codex" {
		t.Errorf("agent = %q, want codex (default)", agent)
	}
}

func TestParseProjectFromReplyNumber(t *testing.T) {
	r := NewResolver([]string{"backend", "frontend"}, nil)

	project, found := r.ParseProjectFromReply("1")
	if !found || project != "backend" {
		t.Errorf("expected backend, got %q (found=%v)", project, found)
	}

	project, found = r.ParseProjectFromReply("2")
	if !found || project != "frontend" {
		t.Errorf("expected frontend, got %q (found=%v)", project, found)
	}
}

func TestParseProjectFromReplyName(t *testing.T) {
	r := NewResolver([]string{"backend", "frontend"}, nil)

	project, found := r.ParseProjectFromReply("backend")
	if !found || project != "backend" {
		t.Errorf("expected backend, got %q (found=%v)", project, found)
	}
}

func TestParseProjectFromReplyNotFound(t *testing.T) {
	r := NewResolver([]string{"backend", "frontend"}, nil)

	_, found := r.ParseProjectFromReply("nonexistent")
	if found {
		t.Error("should not find nonexistent project")
	}
}
