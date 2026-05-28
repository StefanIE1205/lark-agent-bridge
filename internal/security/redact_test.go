package security

import (
	"strings"
	"testing"
)

func TestRedactEnvVarStyle(t *testing.T) {
	tests := []struct {
		input    string
		contains string // what the output should contain
		absent   string // what the output should NOT contain
	}{
		{"OPENAI_API_KEY=sk-abc123", "OPENAI_API_KEY=***", "sk-abc123"},
		{"export API_KEY=secret123", "API_KEY=***", "secret123"},
		{"TOKEN=my-token-value", "TOKEN=***", "my-token-value"},
		{"SECRET=mysecret", "SECRET=***", "mysecret"},
		{"PASSWORD=hunter2", "PASSWORD=***", "hunter2"},
		{"CREDENTIAL=cred123", "CREDENTIAL=***", "cred123"},
		{"MY_API_KEY=abc", "MY_API_KEY=***", "abc"},
		{"APIKEY=xyz", "APIKEY=***", "xyz"},
	}

	for _, tc := range tests {
		result := Redact(tc.input)
		if !strings.Contains(result, tc.contains) {
			t.Errorf("Redact(%q) should contain %q, got: %q", tc.input, tc.contains, result)
		}
		if strings.Contains(result, tc.absent) {
			t.Errorf("Redact(%q) should NOT contain %q, got: %q", tc.input, tc.absent, result)
		}
	}
}

func TestRedactJSONStyle(t *testing.T) {
	tests := []struct {
		input  string
		absent string
	}{
		{`{"api_key": "sk-abc"}`, "sk-abc"},
		{`{"auth_token": "mytoken123"}`, "mytoken123"},
		{`{"secret": "x"}`, "x"},
		{`{"password": "hunter2"}`, "hunter2"},
	}

	for _, tc := range tests {
		result := Redact(tc.input)
		if strings.Contains(result, tc.absent) {
			t.Errorf("Redact(%q) should NOT contain %q, got: %q", tc.input, tc.absent, result)
		}
		if !strings.Contains(result, "***") {
			t.Errorf("Redact(%q) should contain ***, got: %q", tc.input, result)
		}
	}
}

func TestRedactBearer(t *testing.T) {
	input := "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
	result := Redact(input)
	if strings.Contains(result, "eyJhbGci") {
		t.Errorf("Bearer token should be redacted: %q", result)
	}
	if !strings.Contains(result, "Bearer ***") {
		t.Errorf("expected 'Bearer ***' in: %q", result)
	}
}

func TestRedactCLIFlag(t *testing.T) {
	input := "some-command --api-key sk-abc123 --other-arg value"
	result := Redact(input)
	if strings.Contains(result, "sk-abc123") {
		t.Errorf("CLI flag value should be redacted: %q", result)
	}
}

func TestRedactEmpty(t *testing.T) {
	result := Redact("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestRedactNoSecrets(t *testing.T) {
	input := "hello world, this is a normal message"
	result := Redact(input)
	if result != input {
		t.Errorf("expected no change, got: %q", result)
	}
}

func TestRedactCaseInsensitive(t *testing.T) {
	tests := []string{
		"Api_Key=value123",
		"API_KEY=value123",
		"api_key=value123",
	}
	for _, tc := range tests {
		result := Redact(tc)
		if strings.Contains(result, "value123") {
			t.Errorf("case insensitive redact failed for: %q → %q", tc, result)
		}
	}
}

func TestRedactEnv(t *testing.T) {
	input := []string{
		"PATH=/usr/bin",
		"OPENAI_API_KEY=sk-secret-value",
		"HOME=/home/user",
		"GITHUB_TOKEN=ghp_token123",
	}
	result := RedactEnv(input)

	if result[0] != "PATH=/usr/bin" {
		t.Errorf("non-secret should be unchanged: %q", result[0])
	}
	if result[1] != "OPENAI_API_KEY=***" {
		t.Errorf("secret should be redacted: %q", result[1])
	}
	if result[2] != "HOME=/home/user" {
		t.Errorf("non-secret should be unchanged: %q", result[2])
	}
	if result[3] != "GITHUB_TOKEN=***" {
		t.Errorf("secret should be redacted: %q", result[3])
	}
}

func TestRedactEnvNoEquals(t *testing.T) {
	input := []string{"PLAIN_VALUE"}
	result := RedactEnv(input)
	if result[0] != "PLAIN_VALUE" {
		t.Errorf("value without = should be unchanged: %q", result[0])
	}
}

func TestRedactMultiline(t *testing.T) {
	input := "Line 1: hello\nOPENAI_API_KEY=sk-abc\nLine 3: world\nPASSWORD=secret123"
	result := Redact(input)
	if strings.Contains(result, "sk-abc") {
		t.Error("multiline secret should be redacted")
	}
	if strings.Contains(result, "secret123") {
		t.Error("multiline password should be redacted")
	}
	if !strings.Contains(result, "hello") || !strings.Contains(result, "world") {
		t.Error("non-secret lines should be preserved")
	}
}
