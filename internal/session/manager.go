package session

import (
	"fmt"
	"sync"
	"time"
)

type Manager struct {
	mu       sync.Mutex
	sessions map[string]*Session
}

type Session struct {
	State
	cancelFunc func()
}

func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
	}
}

// GetOrCreate returns an existing session or creates a new one.
func (m *Manager) GetOrCreate(chatID, threadID, project, agent, workDir string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := DeriveKey(chatID, threadID, project, agent)
	k := key.String()

	sess, exists := m.sessions[k]
	if exists && sess.Status != StatusClosed {
		sess.LastActivity = time.Now()
		return sess, nil
	}

	sess = &Session{
		State: State{
			Key:          k,
			ChatID:       chatID,
			ThreadID:     threadID,
			Project:      project,
			Agent:        agent,
			WorkDir:      workDir,
			Status:       StatusIdle,
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		},
	}
	m.sessions[k] = sess
	return sess, nil
}

// Get returns a session by its key string, or nil.
func (m *Manager) Get(key string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessions[key]
}

// ListActive returns all non-closed sessions.
func (m *Manager) ListActive() []State {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []State
	for _, s := range m.sessions {
		if s.Status != StatusClosed {
			result = append(result, s.State)
		}
	}
	return result
}

// StartTask transitions a session to running if it's idle.
// Returns an error if the session is busy or in a non-runnable state.
func (m *Manager) StartTask(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s := m.sessions[key]
	if s == nil {
		return fmt.Errorf("session: %s not found", key)
	}

	if s.Status != StatusIdle && s.Status != StatusFailed {
		return fmt.Errorf("session: %s is %s, cannot start task", key, s.Status)
	}

	// Check project-level concurrency: same project can't have two writing tasks
	if s.Status == StatusRunning {
		return fmt.Errorf("session: project %s is already running", s.Project)
	}

	s.Status = StatusRunning
	s.LastActivity = time.Now()
	return nil
}

// Transition attempts to change a session's status.
func (m *Manager) Transition(key string, to Status) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s := m.sessions[key]
	if s == nil {
		return fmt.Errorf("session: %s not found", key)
	}

	if !CanTransition(s.Status, to) {
		return fmt.Errorf("session: cannot transition from %s to %s", s.Status, to)
	}

	s.Status = to
	s.LastActivity = time.Now()
	return nil
}

// Stop cancels the session's context and transitions to stopping.
func (m *Manager) Stop(key string) error {
	m.mu.Lock()
	s := m.sessions[key]
	if s == nil {
		m.mu.Unlock()
		return fmt.Errorf("session: %s not found", key)
	}

	if s.Status != StatusRunning && s.Status != StatusStarting {
		m.mu.Unlock()
		return fmt.Errorf("session: %s is %s, nothing to stop", key, s.Status)
	}

	s.Status = StatusStopping
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
	m.mu.Unlock()

	// Wait briefly for graceful stop, then force idle
	time.Sleep(100 * time.Millisecond)

	m.mu.Lock()
	if s.Status == StatusStopping {
		s.Status = StatusIdle
	}
	m.mu.Unlock()
	return nil
}

// SetCancelFunc stores the cancel function for a session.
func (m *Manager) SetCancelFunc(key string, cancel func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s := m.sessions[key]; s != nil {
		s.cancelFunc = cancel
	}
}

// SetError records an error on a session.
func (m *Manager) SetError(key string, errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s := m.sessions[key]; s != nil {
		s.LastError = errMsg
		s.LastActivity = time.Now()
	}
}

// IsProjectRunning checks if any session for the given project is in a running/starting state.
func (m *Manager) IsProjectRunning(project string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		if s.Project == project && (s.Status == StatusRunning || s.Status == StatusStarting) {
			return true
		}
	}
	return false
}
