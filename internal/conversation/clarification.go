package conversation

import (
	"sync"
	"time"
)

// ClarificationType represents the type of clarification needed.
type ClarificationType string

const (
	ClarifyProject ClarificationType = "project"
	ClarifyAgent   ClarificationType = "agent"
	ClarifyTask    ClarificationType = "task"
)

// Clarification represents a pending question to the user.
type Clarification struct {
	ID        string            `json:"id"`
	ChatID    string            `json:"chat_id"`
	ThreadID  string            `json:"thread_id"`
	Type      ClarificationType `json:"type"`
	Question  string            `json:"question"`
	Options   []string          `json:"options"`
	Task      string            `json:"task"`
	CreatedAt time.Time         `json:"created_at"`
	ExpiresAt time.Time         `json:"expires_at"`
}

// ClarificationManager manages pending clarifications.
type ClarificationManager struct {
	mu             sync.Mutex
	pending        map[string]*Clarification // key: chatID or chatID:threadID
	timeout        time.Duration
}

// NewClarificationManager creates a new clarification manager.
func NewClarificationManager(timeout time.Duration) *ClarificationManager {
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	return &ClarificationManager{
		pending: make(map[string]*Clarification),
		timeout: timeout,
	}
}

// clarificationKey creates a composite key for chat/thread.
func clarificationKey(chatID, threadID string) string {
	if threadID == "" {
		return chatID
	}
	return chatID + ":" + threadID
}

// Save saves a pending clarification.
func (m *ClarificationManager) Save(chatID, threadID string, clarifyType ClarificationType, question string, options []string, task string) *Clarification {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	c := &Clarification{
		ChatID:    chatID,
		ThreadID:  threadID,
		Type:      clarifyType,
		Question:  question,
		Options:   options,
		Task:      task,
		CreatedAt: now,
		ExpiresAt: now.Add(m.timeout),
	}

	key := clarificationKey(chatID, threadID)
	m.pending[key] = c
	return c
}

// Get returns the pending clarification for a chat/thread, or nil.
func (m *ClarificationManager) Get(chatID, threadID string) *Clarification {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := clarificationKey(chatID, threadID)
	c := m.pending[key]
	if c == nil {
		return nil
	}

	// Check expiration
	if time.Now().After(c.ExpiresAt) {
		delete(m.pending, key)
		return nil
	}

	return c
}

// Clear removes the pending clarification for a chat/thread.
func (m *ClarificationManager) Clear(chatID, threadID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := clarificationKey(chatID, threadID)
	delete(m.pending, key)
}

// HasPending checks if there's a pending clarification for a chat/thread.
func (m *ClarificationManager) HasPending(chatID, threadID string) bool {
	return m.Get(chatID, threadID) != nil
}
