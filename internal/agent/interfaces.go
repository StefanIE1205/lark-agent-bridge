package agent

import (
	"context"
	"time"
)

type AgentEventType string

const (
	EventTextDelta         AgentEventType = "text.delta"
	EventToolStart         AgentEventType = "tool.start"
	EventToolOutput        AgentEventType = "tool.output"
	EventApprovalRequested AgentEventType = "approval.requested"
	EventDone              AgentEventType = "done"
	EventError             AgentEventType = "error"
)

type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

type AgentEvent struct {
	Type      AgentEventType `json:"type"`
	Text      string         `json:"text"`
	ToolName  string         `json:"tool_name,omitempty"`
	ToolInput string         `json:"tool_input,omitempty"`
	Error     string         `json:"error,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
	Risk      RiskLevel      `json:"risk,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type SessionOptions struct {
	SessionKey string
	WorkDir    string
	Model      string
	Mode       string
	Env        []string
}

// Capability describes what an agent supports.
type Capability struct {
	PersistentSession bool `json:"persistent_session"`
	SupportsApproval  bool `json:"supports_approval"`
	SupportsStreaming bool `json:"supports_streaming"`
	Experimental      bool `json:"experimental"`
}

type Agent interface {
	Name() string
	StartSession(ctx context.Context, opts SessionOptions) (AgentSession, error)
	Check(ctx context.Context) error
	Capabilities() Capability
}

type AgentSession interface {
	Send(ctx context.Context, prompt string) error
	Events() <-chan AgentEvent
	Stop(ctx context.Context) error
	Alive() bool
}
