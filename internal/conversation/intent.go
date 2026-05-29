package conversation

import (
	"strings"
)

// IntentType represents the type of user intent.
type IntentType string

const (
	IntentAsk           IntentType = "ask"
	IntentStatus        IntentType = "status"
	IntentStop          IntentType = "stop"
	IntentSwitchProject IntentType = "switch_project"
	IntentSwitchAgent   IntentType = "switch_agent"
	IntentApprove       IntentType = "approve"
	IntentDeny          IntentType = "deny"
	IntentHelp          IntentType = "help"
	IntentIgnore        IntentType = "ignore"
	IntentClarifyReply  IntentType = "clarify_reply"
)

// Intent represents a parsed user intent.
type Intent struct {
	Type        IntentType
	Task        string
	ProjectHint string
	AgentHint   string
	Confidence  float64
	Raw         string
}

// Router parses natural language text into intents.
type Router struct {
	projectNames []string
	agentNames   []string
}

// NewRouter creates a new intent router.
func NewRouter(projectNames, agentNames []string) *Router {
	return &Router{
		projectNames: projectNames,
		agentNames:   agentNames,
	}
}

// Parse analyzes the input text and returns an Intent.
func (r *Router) Parse(text string) Intent {
	text = strings.TrimSpace(text)
	if text == "" {
		return Intent{Type: IntentIgnore, Raw: text}
	}

	// Check for stop intent
	if r.isStop(text) {
		return Intent{Type: IntentStop, Raw: text}
	}

	// Check for status intent
	if r.isStatus(text) {
		return Intent{Type: IntentStatus, Raw: text}
	}

	// Check for help intent
	if r.isHelp(text) {
		return Intent{Type: IntentHelp, Raw: text}
	}

	// Check for approve/deny
	if r.isApprove(text) {
		return Intent{Type: IntentApprove, Raw: text}
	}
	if r.isDeny(text) {
		return Intent{Type: IntentDeny, Raw: text}
	}

	// Check for agent hint
	agentHint := r.extractAgentHint(text)

	// Check for project hint
	projectHint := r.extractProjectHint(text)

	// Default: treat as a task (ask)
	task := text
	return Intent{
		Type:        IntentAsk,
		Task:        task,
		ProjectHint: projectHint,
		AgentHint:   agentHint,
		Confidence:  0.8,
		Raw:         text,
	}
}

func (r *Router) isStop(text string) bool {
	stopWords := []string{"停一下", "停止", "别做了", "暂停", "停下", "stop", "cancel"}
	lower := strings.ToLower(text)
	for _, w := range stopWords {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}

func (r *Router) isStatus(text string) bool {
	statusWords := []string{"状态", "进度", "做到哪", "什么情况", "怎么样", "status"}
	lower := strings.ToLower(text)
	for _, w := range statusWords {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}

func (r *Router) isHelp(text string) bool {
	helpWords := []string{"帮助", "怎么用", "你会什么", "help"}
	lower := strings.ToLower(text)
	for _, w := range helpWords {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}

func (r *Router) isApprove(text string) bool {
	approveWords := []string{"可以", "同意", "批准", "继续", "好的", "行", "approve", "yes", "ok"}
	lower := strings.ToLower(text)
	for _, w := range approveWords {
		if lower == w || strings.HasPrefix(lower, w+" ") {
			return true
		}
	}
	return false
}

func (r *Router) isDeny(text string) bool {
	denyWords := []string{"不行", "拒绝", "不要", "别执行", "deny", "no"}
	lower := strings.ToLower(text)
	for _, w := range denyWords {
		if lower == w || strings.HasPrefix(lower, w+" ") {
			return true
		}
	}
	return false
}

func (r *Router) extractAgentHint(text string) string {
	lower := strings.ToLower(text)
	for _, name := range r.agentNames {
		if strings.Contains(lower, strings.ToLower(name)) {
			return name
		}
	}
	// Common aliases
	agentAliases := map[string]string{
		"claude":     "claude",
		"codex":      "codex",
		"antigravity": "antigravity",
	}
	for alias, name := range agentAliases {
		if strings.Contains(lower, alias) {
			return name
		}
	}
	return ""
}

func (r *Router) extractProjectHint(text string) string {
	lower := strings.ToLower(text)
	for _, name := range r.projectNames {
		if strings.Contains(lower, strings.ToLower(name)) {
			return name
		}
	}
	return ""
}
