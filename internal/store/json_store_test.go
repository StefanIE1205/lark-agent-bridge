package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteAndRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	data := map[string]string{"a": "1", "b": "2"}
	if err := AtomicWriteJSON(path, data); err != nil {
		t.Fatalf("AtomicWriteJSON: %v", err)
	}

	var out map[string]string
	if err := ReadJSON(path, &out); err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	if len(out) != 2 || out["a"] != "1" {
		t.Errorf("data mismatch: %v", out)
	}
}

func TestReadJSONFileNotFound(t *testing.T) {
	var m map[string]string
	err := ReadJSON("/nonexistent/path.json", &m)
	if err != nil {
		t.Fatalf("ReadJSON should return nil for missing file, got: %v", err)
	}
	if m != nil {
		t.Errorf("expected nil map for missing file, got: %v", m)
	}
}

func TestAtomicWriteOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	_ = AtomicWriteJSON(path, map[string]string{"first": "write"})
	_ = AtomicWriteJSON(path, map[string]string{"second": "write"})

	var out map[string]string
	_ = ReadJSON(path, &out)
	if out["second"] != "write" || len(out) != 1 {
		t.Errorf("expected single key 'second', got: %v", out)
	}
}

func TestStateStoreBindAndLoad(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStateStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.BindProject("demo", "C:\\demo"); err != nil {
		t.Fatal(err)
	}
	if err := s.BindProject("backend", "D:\\backend"); err != nil {
		t.Fatal(err)
	}

	projects, err := s.LoadProjects()
	if err != nil {
		t.Fatal(err)
	}
	if projects["demo"] != "C:\\demo" {
		t.Errorf("demo = %q", projects["demo"])
	}
	if projects["backend"] != "D:\\backend" {
		t.Errorf("backend = %q", projects["backend"])
	}
}

func TestStateStoreSetAndGetChatDefaults(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStateStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.SetChatDefault("oc_chat", "project", "demo"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetChatDefault("oc_chat", "agent", "codex"); err != nil {
		t.Fatal(err)
	}

	defaults, err := s.GetChatDefaults("oc_chat")
	if err != nil {
		t.Fatal(err)
	}
	if defaults.Project != "demo" {
		t.Errorf("project = %q, want demo", defaults.Project)
	}
	if defaults.Agent != "codex" {
		t.Errorf("agent = %q, want codex", defaults.Agent)
	}
}

func TestStateStoreChatDefaultsPersist(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStateStore(dir)

	s.SetChatDefault("oc_chat", "project", "demo")
	s.SetChatDefault("oc_chat", "agent", "claude")

	// Create a new store pointing to same dir — should reload
	s2, _ := NewStateStore(dir)
	defaults, _ := s2.GetChatDefaults("oc_chat")

	if defaults.Project != "demo" || defaults.Agent != "claude" {
		t.Errorf("persistence broken: %+v", defaults)
	}
}

func TestStateStoreGetMissingChat(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStateStore(dir)

	defaults, err := s.GetChatDefaults("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if defaults.Project != "" || defaults.Agent != "" {
		t.Errorf("expected empty defaults, got: %+v", defaults)
	}
}

func TestStateStoreFileCreated(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStateStore(dir)

	s.BindProject("test", "/tmp/test")

	if _, err := os.Stat(s.projectsPath()); err != nil {
		t.Errorf("projects.json should exist: %v", err)
	}
}
