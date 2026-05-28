package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type ApprovalStatus string

const (
	ApprovalPending ApprovalStatus = "pending"
	ApprovalGranted ApprovalStatus = "granted"
	ApprovalDenied  ApprovalStatus = "denied"
	ApprovalExpired ApprovalStatus = "expired"
)

type ApprovalRequest struct {
	ID         string         `json:"id"`
	SessionKey string         `json:"session_key"`
	Action     string         `json:"action"`
	Risk       string         `json:"risk"`
	Status     ApprovalStatus `json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
	ExpiresAt  time.Time      `json:"expires_at"`
	ResolvedAt *time.Time     `json:"resolved_at,omitempty"`
}

type ApprovalManager struct {
	mu       sync.Mutex
	requests map[string]*ApprovalRequest
	timeout  time.Duration
}

func NewApprovalManager(timeout time.Duration) *ApprovalManager {
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	return &ApprovalManager{
		requests: make(map[string]*ApprovalRequest),
		timeout:  timeout,
	}
}

func (m *ApprovalManager) Create(sessionKey, action, risk string) *ApprovalRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	req := &ApprovalRequest{
		ID:         newID(),
		SessionKey: sessionKey,
		Action:     action,
		Risk:       risk,
		Status:     ApprovalPending,
		CreatedAt:  now,
		ExpiresAt:  now.Add(m.timeout),
	}

	m.requests[req.ID] = req
	return req
}

func (m *ApprovalManager) Approve(id string) error {
	return m.resolve(id, ApprovalGranted)
}

func (m *ApprovalManager) Deny(id string) error {
	return m.resolve(id, ApprovalDenied)
}

func (m *ApprovalManager) resolve(id string, status ApprovalStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	req, ok := m.requests[id]
	if !ok {
		return fmt.Errorf("approval: request %s not found", id)
	}

	if req.Status != ApprovalPending {
		return fmt.Errorf("approval: request %s is already %s", id, req.Status)
	}

	if time.Now().After(req.ExpiresAt) {
		req.Status = ApprovalExpired
		return fmt.Errorf("approval: request %s has expired", id)
	}

	req.Status = status
	now := time.Now()
	req.ResolvedAt = &now
	return nil
}

func (m *ApprovalManager) Get(id string) *ApprovalRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.requests[id]
}

func (m *ApprovalManager) Pending() []ApprovalRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cleanupLocked()

	var result []ApprovalRequest
	for _, r := range m.requests {
		if r.Status == ApprovalPending {
			result = append(result, *r)
		}
	}
	return result
}

func (m *ApprovalManager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanupLocked()
}

func newID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (m *ApprovalManager) cleanupLocked() {
	now := time.Now()
	for id, r := range m.requests {
		if r.Status == ApprovalPending && now.After(r.ExpiresAt) {
			r.Status = ApprovalExpired
			now := time.Now()
			r.ResolvedAt = &now
		}
		// Remove old resolved requests
		if r.Status != ApprovalPending && r.ResolvedAt != nil && now.Sub(*r.ResolvedAt) > m.timeout {
			delete(m.requests, id)
		}
	}
}
