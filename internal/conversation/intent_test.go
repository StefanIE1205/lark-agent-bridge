package conversation

import (
	"testing"
)

func TestParseEmpty(t *testing.T) {
	r := NewRouter(nil, nil)
	intent := r.Parse("")
	if intent.Type != IntentIgnore {
		t.Errorf("empty text should be ignore, got %s", intent.Type)
	}
}

func TestParseStop(t *testing.T) {
	r := NewRouter(nil, nil)
	tests := []string{"停一下", "停止", "别做了", "stop", "Stop"}
	for _, text := range tests {
		intent := r.Parse(text)
		if intent.Type != IntentStop {
			t.Errorf("%q should be stop, got %s", text, intent.Type)
		}
	}
}

func TestParseStatus(t *testing.T) {
	r := NewRouter(nil, nil)
	tests := []string{"状态", "进度", "做到哪了", "status"}
	for _, text := range tests {
		intent := r.Parse(text)
		if intent.Type != IntentStatus {
			t.Errorf("%q should be status, got %s", text, intent.Type)
		}
	}
}

func TestParseHelp(t *testing.T) {
	r := NewRouter(nil, nil)
	tests := []string{"帮助", "怎么用", "help"}
	for _, text := range tests {
		intent := r.Parse(text)
		if intent.Type != IntentHelp {
			t.Errorf("%q should be help, got %s", text, intent.Type)
		}
	}
}

func TestParseApprove(t *testing.T) {
	r := NewRouter(nil, nil)
	tests := []string{"可以", "同意", "继续", "好的", "approve"}
	for _, text := range tests {
		intent := r.Parse(text)
		if intent.Type != IntentApprove {
			t.Errorf("%q should be approve, got %s", text, intent.Type)
		}
	}
}

func TestParseDeny(t *testing.T) {
	r := NewRouter(nil, nil)
	tests := []string{"不行", "拒绝", "不要", "deny"}
	for _, text := range tests {
		intent := r.Parse(text)
		if intent.Type != IntentDeny {
			t.Errorf("%q should be deny, got %s", text, intent.Type)
		}
	}
}

func TestParseAsk(t *testing.T) {
	r := NewRouter(nil, nil)
	tests := []string{
		"帮我跑一下测试",
		"看看为什么失败",
		"修复这个 bug",
		"解释一下这个模块",
	}
	for _, text := range tests {
		intent := r.Parse(text)
		if intent.Type != IntentAsk {
			t.Errorf("%q should be ask, got %s", text, intent.Type)
		}
	}
}

func TestParseAgentHint(t *testing.T) {
	r := NewRouter(nil, []string{"claude", "codex"})

	tests := []struct {
		text string
		want string
	}{
		{"用 Claude 看一下", "claude"},
		{"让 Codex 跑一下", "codex"},
		{"帮我修复 bug", ""},
	}

	for _, tt := range tests {
		intent := r.Parse(tt.text)
		if intent.AgentHint != tt.want {
			t.Errorf("%q: agent hint = %q, want %q", tt.text, intent.AgentHint, tt.want)
		}
	}
}

func TestParseProjectHint(t *testing.T) {
	r := NewRouter([]string{"backend", "frontend"}, nil)

	tests := []struct {
		text string
		want string
	}{
		{"帮我看 backend 项目", "backend"},
		{"检查 frontend 代码", "frontend"},
		{"帮我修复 bug", ""},
	}

	for _, tt := range tests {
		intent := r.Parse(tt.text)
		if intent.ProjectHint != tt.want {
			t.Errorf("%q: project hint = %q, want %q", tt.text, intent.ProjectHint, tt.want)
		}
	}
}
