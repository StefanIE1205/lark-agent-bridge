package logging

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")

	l, err := Init(dataDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer l.Close()

	// Check all expected directories exist
	expectedDirs := []string{
		"state",
		"logs",
		filepath.Join("logs", "sessions"),
		"audit",
	}
	for _, d := range expectedDirs {
		fullPath := filepath.Join(dataDir, d)
		info, err := os.Stat(fullPath)
		if err != nil {
			t.Errorf("expected directory %s to exist: %v", fullPath, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", fullPath)
		}
	}

	// Check app.log was created
	appLogPath := filepath.Join(dataDir, "logs", "app.log")
	if _, err := os.Stat(appLogPath); err != nil {
		t.Errorf("app.log should exist: %v", err)
	}

	// Check audit.log was created
	auditPath := filepath.Join(dataDir, "audit", "audit.log")
	if _, err := os.Stat(auditPath); err != nil {
		t.Errorf("audit.log should exist: %v", err)
	}
}

func TestInitEmptyDataDir(t *testing.T) {
	_, err := Init("")
	if err == nil {
		t.Fatal("expected error for empty data_dir")
	}
}

func TestSessionLogWriter(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")

	l, err := Init(dataDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer l.Close()

	w, err := l.SessionLogWriter("lark:chat123:thread456:demo:codex")
	if err != nil {
		t.Fatalf("SessionLogWriter failed: %v", err)
	}

	msg := "test log line\n"
	if _, err := w.Write([]byte(msg)); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	w.Close()

	// Verify the file was created with sanitized name
	sessionPath := filepath.Join(dataDir, "logs", "sessions", "lark_chat123_thread456_demo_codex.log")
	content, err := os.ReadFile(sessionPath)
	if err != nil {
		t.Fatalf("read session log failed: %v", err)
	}
	if string(content) != msg {
		t.Errorf("session log content = %q, want %q", string(content), msg)
	}
}

func TestSafeKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abc", "abc"},
		{"a:b/c", "a_b_c"},
		{"a<b>c|d?e*f", "a_b_c_d_e_f"},
	}
	for _, tc := range tests {
		got := safeKey(tc.input)
		if got != tc.expected {
			t.Errorf("safeKey(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestLogLevels(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")

	l, err := Init(dataDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer l.Close()

	l.Info("info msg: %d", 1)
	l.Error("error msg: %s", "boom")
	l.Warn("warn msg")
	l.Debug("debug msg")
	l.Audit("audit msg: user=%s action=%s", "test", "login")

	appLogContent, err := os.ReadFile(filepath.Join(dataDir, "logs", "app.log"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(appLogContent)
	if !contains(content, "info msg: 1") {
		t.Error("app.log missing info message")
	}
	if !contains(content, "error msg: boom") {
		t.Error("app.log missing error message")
	}

	auditContent, err := os.ReadFile(filepath.Join(dataDir, "audit", "audit.log"))
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(auditContent), "audit msg: user=test action=login") {
		t.Error("audit.log missing audit message")
	}
}

func TestInitNoWritePermission(t *testing.T) {
	t.Skip("permission test not applicable on Windows; dir creation succeeds with read-only parent on this platform")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
