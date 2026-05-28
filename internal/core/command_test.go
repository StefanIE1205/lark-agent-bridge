package core

import (
	"reflect"
	"testing"
)

func TestParsePing(t *testing.T) {
	c := ParseCommand("/ping")
	if c.Name != "ping" {
		t.Errorf("name = %q, want ping", c.Name)
	}
}

func TestParseHelp(t *testing.T) {
	c := ParseCommand("/help")
	if c.Name != "help" {
		t.Errorf("name = %q, want help", c.Name)
	}
}

func TestParseAsk(t *testing.T) {
	c := ParseCommand("/ask fix the bug in parser.go")
	if c.Name != "ask" {
		t.Errorf("name = %q, want ask", c.Name)
	}
	if len(c.Args) != 1 {
		t.Fatalf("args count = %d, want 1", len(c.Args))
	}
	if c.Args[0] != "fix the bug in parser.go" {
		t.Errorf("args[0] = %q, want %q", c.Args[0], "fix the bug in parser.go")
	}
}

func TestParseAskEmpty(t *testing.T) {
	c := ParseCommand("/ask")
	if c.Name != "ask" {
		t.Errorf("name = %q, want ask", c.Name)
	}
}

func TestParseBangShorthand(t *testing.T) {
	c := ParseCommand("! fix the bug")
	if c.Name != "ask" {
		t.Errorf("name = %q, want ask", c.Name)
	}
	if len(c.Args) != 1 {
		t.Fatalf("args count = %d, want 1", len(c.Args))
	}
	if c.Args[0] != "fix the bug" {
		t.Errorf("args[0] = %q, want %q", c.Args[0], "fix the bug")
	}
}

func TestParseBangEmpty(t *testing.T) {
	c := ParseCommand("!")
	if c.Name != "" {
		t.Errorf("expected empty command for bare !, got %q", c.Name)
	}
}

func TestParseBind(t *testing.T) {
	c := ParseCommand(`/bind demo C:\repo\demo`)
	if c.Name != "bind" {
		t.Errorf("name = %q, want bind", c.Name)
	}
	if len(c.Args) != 2 {
		t.Fatalf("args count = %d, want 2", len(c.Args))
	}
	if c.Args[0] != "demo" {
		t.Errorf("args[0] = %q, want demo", c.Args[0])
	}
	if c.Args[1] != `C:\repo\demo` {
		t.Errorf("args[1] = %q, want C:\\repo\\demo", c.Args[1])
	}
}

func TestParseBindQuotedPath(t *testing.T) {
	c := ParseCommand(`/bind demo "C:\My Repo"`)
	if c.Name != "bind" {
		t.Errorf("name = %q, want bind", c.Name)
	}
	if len(c.Args) != 2 {
		t.Fatalf("args count = %d, want 2", len(c.Args))
	}
	if c.Args[1] != `C:\My Repo` {
		t.Errorf("args[1] = %q, want %q", c.Args[1], `C:\My Repo`)
	}
}

func TestParseBindNoPath(t *testing.T) {
	c := ParseCommand("/bind demo")
	if c.Name != "bind" {
		t.Errorf("name = %q, want bind", c.Name)
	}
	if len(c.Args) != 1 {
		t.Fatalf("args count = %d, want 1", len(c.Args))
	}
	if c.Args[0] != "demo" {
		t.Errorf("args[0] = %q, want demo", c.Args[0])
	}
}

func TestParseProject(t *testing.T) {
	c := ParseCommand("/project myproject")
	if c.Name != "project" {
		t.Errorf("name = %q, want project", c.Name)
	}
	if len(c.Args) != 1 || c.Args[0] != "myproject" {
		t.Errorf("args = %v, want [myproject]", c.Args)
	}
}

func TestParseAgent(t *testing.T) {
	c := ParseCommand("/agent codex")
	if c.Name != "agent" {
		t.Errorf("name = %q, want agent", c.Name)
	}
	if len(c.Args) != 1 || c.Args[0] != "codex" {
		t.Errorf("args = %v, want [codex]", c.Args)
	}
}

func TestParseAgentClaude(t *testing.T) {
	c := ParseCommand("/agent claude")
	if c.Name != "agent" || c.Args[0] != "claude" {
		t.Errorf("got name=%s args=%v, want name=agent args=[claude]", c.Name, c.Args)
	}
}

func TestParseStatus(t *testing.T) {
	c := ParseCommand("/status")
	if c.Name != "status" {
		t.Errorf("name = %q, want status", c.Name)
	}
}

func TestParseSessions(t *testing.T) {
	c := ParseCommand("/sessions")
	if c.Name != "sessions" {
		t.Errorf("name = %q, want sessions", c.Name)
	}
}

func TestParseStop(t *testing.T) {
	c := ParseCommand("/stop")
	if c.Name != "stop" {
		t.Errorf("name = %q, want stop", c.Name)
	}
}

func TestParseLogDefault(t *testing.T) {
	c := ParseCommand("/log")
	if c.Name != "log" {
		t.Errorf("name = %q, want log", c.Name)
	}
	if len(c.Args) != 0 {
		t.Errorf("expected no args, got %v", c.Args)
	}
}

func TestParseLogWithNumber(t *testing.T) {
	c := ParseCommand("/log 100")
	if c.Name != "log" || len(c.Args) != 1 || c.Args[0] != "100" {
		t.Errorf("got name=%s args=%v, want name=log args=[100]", c.Name, c.Args)
	}
}

func TestParseApprove(t *testing.T) {
	c := ParseCommand("/approve req-001")
	if c.Name != "approve" || len(c.Args) != 1 || c.Args[0] != "req-001" {
		t.Errorf("got name=%s args=%v, want name=approve args=[req-001]", c.Name, c.Args)
	}
}

func TestParseDeny(t *testing.T) {
	c := ParseCommand("/deny req-002")
	if c.Name != "deny" || len(c.Args) != 1 || c.Args[0] != "req-002" {
		t.Errorf("got name=%s args=%v, want name=deny args=[req-002]", c.Name, c.Args)
	}
}

func TestParseEmptyMessage(t *testing.T) {
	c := ParseCommand("")
	if c.Name != "" {
		t.Errorf("expected empty command, got %q", c.Name)
	}
}

func TestParseWhitespaceOnly(t *testing.T) {
	c := ParseCommand("   ")
	if c.Name != "" {
		t.Errorf("expected empty command, got %q", c.Name)
	}
}

func TestParseFullWidthSlash(t *testing.T) {
	c := ParseCommand("／help")
	if c.Name != "help" {
		t.Errorf("name = %q, want help (full-width slash should be converted)", c.Name)
	}
}

func TestParseFullWidthExclamation(t *testing.T) {
	c := ParseCommand("！fix bug")
	if c.Name != "ask" {
		t.Errorf("name = %q, want ask (full-width ! should be converted)", c.Name)
	}
	if c.Args[0] != "fix bug" {
		t.Errorf("args = %v, want [fix bug]", c.Args)
	}
}

func TestParseNonCommand(t *testing.T) {
	c := ParseCommand("hello world")
	if c.Name != "" {
		t.Errorf("expected empty command for non-command text, got %q", c.Name)
	}
}

func TestParseUnknownCommand(t *testing.T) {
	c := ParseCommand("/unknown thing")
	if c.Name != "unknown" {
		t.Errorf("name = %q, want unknown", c.Name)
	}
}

func TestParseCaseInsensitive(t *testing.T) {
	c := ParseCommand("/PING")
	if c.Name != "ping" {
		t.Errorf("name = %q, want ping (should be lowercased)", c.Name)
	}
}

func TestCommandRoundTrip(t *testing.T) {
	tests := []struct {
		input    string
		expected Command
	}{
		{"/ping", Command{Name: "ping"}},
		{"/help", Command{Name: "help"}},
		{"/status", Command{Name: "status"}},
		{"/sessions", Command{Name: "sessions"}},
		{"/stop", Command{Name: "stop"}},
		{"/ask fix bug", Command{Name: "ask", Args: []string{"fix bug"}}},
		{"! fix bug", Command{Name: "ask", Args: []string{"fix bug"}}},
		{"/agent codex", Command{Name: "agent", Args: []string{"codex"}}},
		{"/project demo", Command{Name: "project", Args: []string{"demo"}}},
		{"/bind demo C:\\x", Command{Name: "bind", Args: []string{"demo", "C:\\x"}}},
		{`/bind demo "C:\My Repo"`, Command{Name: "bind", Args: []string{"demo", "C:\\My Repo"}}},
		{"/log", Command{Name: "log"}},
		{"/log 50", Command{Name: "log", Args: []string{"50"}}},
		{"/approve req-1", Command{Name: "approve", Args: []string{"req-1"}}},
		{"/deny req-2", Command{Name: "deny", Args: []string{"req-2"}}},
		{"", Command{}},
		{"   ", Command{}},
		{"hello", Command{}},
	}

	for _, tc := range tests {
		c := ParseCommand(tc.input)
		if c.Name != tc.expected.Name {
			t.Errorf("ParseCommand(%q).Name = %q, want %q", tc.input, c.Name, tc.expected.Name)
		}
		if !reflect.DeepEqual(c.Args, tc.expected.Args) {
			t.Errorf("ParseCommand(%q).Args = %v, want %v", tc.input, c.Args, tc.expected.Args)
		}
	}
}
