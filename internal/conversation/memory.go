package conversation

import (
	"sync"
	"time"
)

// Context holds the conversation context for a chat/thread.
type Context struct {
	ChatID         string    `json:"chat_id"`
	ThreadID       string    `json:"thread_id"`
	LastProject    string    `json:"last_project"`
	LastAgent      string    `json:"last_agent"`
	LastTask       string    `json:"last_task"`
	LastTaskStatus string    `json:"last_task_status"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Memory manages conversation contexts.
type Memory struct {
	mu       sync.Mutex
	contexts map[string]*Context
}

// NewMemory creates a new conversation memory.
func NewMemory() *Memory {
	return &Memory{
		contexts: make(map[string]*Context),
	}
}

// contextKey creates a composite key for chat/thread.
func contextKey(chatID, threadID string) string {
	if threadID == "" {
		return chatID
	}
	return chatID + ":" + threadID
}

// Get returns the context for a chat/thread, or a new empty context.
func (m *Memory) Get(chatID, threadID string) Context {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := contextKey(chatID, threadID)
	if ctx, ok := m.contexts[key]; ok {
		return *ctx
	}
	return Context{ChatID: chatID, ThreadID: threadID}
}

// Update updates the context for a chat/thread.
func (m *Memory) Update(chatID, threadID string, fn func(ctx *Context)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := contextKey(chatID, threadID)
	ctx, ok := m.contexts[key]
	if !ok {
		ctx = &Context{ChatID: chatID, ThreadID: threadID}
		m.contexts[key] = ctx
	}
	fn(ctx)
	ctx.UpdatedAt = time.Now()
}

// SetProject updates the last project for a chat/thread.
func (m *Memory) SetProject(chatID, threadID, project string) {
	m.Update(chatID, threadID, func(ctx *Context) {
		ctx.LastProject = project
	})
}

// SetAgent updates the last agent for a chat/thread.
func (m *Memory) SetAgent(chatID, threadID, agent string) {
	m.Update(chatID, threadID, func(ctx *Context) {
		ctx.LastAgent = agent
	})
}

// SetTask updates the last task and status for a chat/thread.
func (m *Memory) SetTask(chatID, threadID, task, status string) {
	m.Update(chatID, threadID, func(ctx *Context) {
		ctx.LastTask = task
		ctx.LastTaskStatus = status
	})
}

// SetTaskStatus updates only the task status.
func (m *Memory) SetTaskStatus(chatID, threadID, status string) {
	m.Update(chatID, threadID, func(ctx *Context) {
		ctx.LastTaskStatus = status
	})
}
