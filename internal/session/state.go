package session

import (
	"fmt"
	"time"
)

type Status string

const (
	StatusIdle            Status = "idle"
	StatusStarting        Status = "starting"
	StatusRunning         Status = "running"
	StatusWaitingApproval Status = "waiting_approval"
	StatusStopping        Status = "stopping"
	StatusFailed          Status = "failed"
	StatusClosed          Status = "closed"
)

type State struct {
	Key          string    `json:"key"`
	ChatID       string    `json:"chat_id"`
	ThreadID     string    `json:"thread_id"`
	Project      string    `json:"project"`
	Agent        string    `json:"agent"`
	WorkDir      string    `json:"work_dir"`
	Status       Status    `json:"status"`
	LastError    string    `json:"last_error"`
	LastActivity time.Time `json:"last_activity"`
	CreatedAt    time.Time `json:"created_at"`
}

type SessionKey struct {
	ChatID   string
	ThreadID string
	Project  string
	Agent    string
}

func (k SessionKey) String() string {
	return fmt.Sprintf("lark:%s:%s:%s:%s", k.ChatID, k.ThreadID, k.Project, k.Agent)
}

func DeriveKey(chatID, threadID, project, agent string) SessionKey {
	if threadID == "" {
		threadID = chatID // fallback to chatID when no thread
	}
	return SessionKey{
		ChatID:   chatID,
		ThreadID: threadID,
		Project:  project,
		Agent:    agent,
	}
}

// ValidTransitions defines which status transitions are allowed.
var ValidTransitions = map[Status][]Status{
	StatusIdle:            {StatusStarting, StatusClosed},
	StatusStarting:        {StatusRunning, StatusFailed},
	StatusRunning:         {StatusWaitingApproval, StatusStopping, StatusFailed, StatusClosed, StatusIdle},
	StatusWaitingApproval: {StatusRunning, StatusStopping, StatusFailed},
	StatusStopping:        {StatusIdle, StatusFailed, StatusClosed},
	StatusFailed:          {StatusIdle, StatusClosed},
	StatusClosed:          {},
}

func CanTransition(from, to Status) bool {
	targets, ok := ValidTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}
