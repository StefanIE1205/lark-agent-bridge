package security

import (
	"testing"
	"time"
)

func TestApprovalCreate(t *testing.T) {
	m := NewApprovalManager(5 * time.Minute)
	req := m.Create("session-1", "write file: main.go", "high")

	if req.ID == "" {
		t.Error("request ID should not be empty")
	}
	if req.SessionKey != "session-1" {
		t.Errorf("SessionKey = %q", req.SessionKey)
	}
	if req.Action != "write file: main.go" {
		t.Errorf("Action = %q", req.Action)
	}
	if req.Risk != "high" {
		t.Errorf("Risk = %q", req.Risk)
	}
	if req.Status != ApprovalPending {
		t.Errorf("Status = %q, want pending", req.Status)
	}
	if req.ExpiresAt.Before(time.Now()) {
		t.Error("ExpiresAt is in the past")
	}
}

func TestApprovalApprove(t *testing.T) {
	m := NewApprovalManager(5 * time.Minute)
	req := m.Create("sess", "install deps", "medium")

	if err := m.Approve(req.ID); err != nil {
		t.Fatalf("Approve failed: %v", err)
	}

	got := m.Get(req.ID)
	if got.Status != ApprovalGranted {
		t.Errorf("Status = %q, want granted", got.Status)
	}
	if got.ResolvedAt == nil {
		t.Error("ResolvedAt should be set")
	}
}

func TestApprovalDeny(t *testing.T) {
	m := NewApprovalManager(5 * time.Minute)
	req := m.Create("sess", "git push", "critical")

	if err := m.Deny(req.ID); err != nil {
		t.Fatalf("Deny failed: %v", err)
	}

	got := m.Get(req.ID)
	if got.Status != ApprovalDenied {
		t.Errorf("Status = %q, want denied", got.Status)
	}
}

func TestApprovalNotFound(t *testing.T) {
	m := NewApprovalManager(5 * time.Minute)
	err := m.Approve("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent request")
	}
}

func TestApprovalDoubleResolve(t *testing.T) {
	m := NewApprovalManager(5 * time.Minute)
	req := m.Create("sess", "test", "low")

	m.Approve(req.ID)
	err := m.Deny(req.ID)
	if err == nil {
		t.Error("expected error when resolving already-resolved request")
	}
}

func TestApprovalExpired(t *testing.T) {
	m := NewApprovalManager(1 * time.Millisecond)
	req := m.Create("sess", "test", "low")

	time.Sleep(10 * time.Millisecond)

	err := m.Approve(req.ID)
	if err == nil {
		t.Error("expected error for expired request")
	}

	got := m.Get(req.ID)
	if got.Status != ApprovalExpired {
		t.Errorf("Status = %q, want expired", got.Status)
	}
}

func TestApprovalPending(t *testing.T) {
	m := NewApprovalManager(5 * time.Minute)
	m.Create("s1", "write file", "high")
	m.Create("s2", "install dep", "medium")
	req3 := m.Create("s3", "git push", "critical")

	m.Approve(req3.ID) // resolve one

	pending := m.Pending()
	if len(pending) != 2 {
		t.Errorf("expected 2 pending, got %d", len(pending))
	}
}

func TestApprovalAutoExpire(t *testing.T) {
	m := NewApprovalManager(1 * time.Millisecond)
	m.Create("s", "test", "low")

	time.Sleep(10 * time.Millisecond)

	// Pending() triggers cleanup
	pending := m.Pending()
	if len(pending) != 0 {
		t.Errorf("expected 0 pending after expiry, got %d", len(pending))
	}
}

func TestApprovalDefaultTimeout(t *testing.T) {
	m := NewApprovalManager(0)
	if m.timeout != 5*time.Minute {
		t.Errorf("default timeout = %v, want 5m", m.timeout)
	}
}

func TestApprovalUniqueIDs(t *testing.T) {
	m := NewApprovalManager(5 * time.Minute)
	r1 := m.Create("s", "test", "low")
	r2 := m.Create("s", "test", "low")

	if r1.ID == r2.ID {
		t.Error("approval IDs should be unique")
	}
}
