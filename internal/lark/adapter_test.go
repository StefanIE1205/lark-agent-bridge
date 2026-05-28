package lark

import (
	"testing"
	"time"

	"github.com/StefanIE1205/lark-agent-bridge/internal/core"
)

func TestNewAdapter(t *testing.T) {
	a := NewAdapter("cli_test", "secret123", nil)
	if a == nil {
		t.Fatal("NewAdapter returned nil")
	}
	if a.Name() != "lark" {
		t.Errorf("Name() = %q, want %q", a.Name(), "lark")
	}
	if a.appID != "cli_test" {
		t.Errorf("appID = %q, want %q", a.appID, "cli_test")
	}
	if a.appSecret != "secret123" {
		t.Errorf("appSecret = %q, want %q", a.appSecret, "secret123")
	}
}

func TestAdapterImplementsPlatform(t *testing.T) {
	a := NewAdapter("test", "secret", nil)
	if _, ok := interface{}(a).(core.Platform); !ok {
		t.Error("Adapter does not implement core.Platform")
	}
}

func TestUpdateNotImplemented(t *testing.T) {
	a := NewAdapter("test", "secret", nil)
	err := a.Update(nil, core.ReplyTarget{}, "msg_1", "updated")
	if err == nil {
		t.Error("Update should return not implemented error")
	}
}

func TestIsDuplicate(t *testing.T) {
	a := NewAdapter("test", "secret", nil)

	if a.isDuplicate("msg_001") {
		t.Error("first occurrence should not be duplicate")
	}
	if !a.isDuplicate("msg_001") {
		t.Error("second occurrence should be duplicate")
	}
	if a.isDuplicate("msg_002") {
		t.Error("different id should not be duplicate")
	}
}

func TestPurgeOldEntries(t *testing.T) {
	a := NewAdapter("test", "secret", nil)

	// Insert an entry with an old timestamp
	a.dedupMu.Lock()
	a.dedup["old_msg"] = dedupEntry{seen: time.Now().Add(-10 * time.Minute)}
	a.dedup["new_msg"] = dedupEntry{seen: time.Now()}
	a.dedupMu.Unlock()

	a.purgeOldEntries()

	a.dedupMu.Lock()
	defer a.dedupMu.Unlock()

	if _, exists := a.dedup["old_msg"]; exists {
		t.Error("old entry should have been purged")
	}
	if _, exists := a.dedup["new_msg"]; !exists {
		t.Error("recent entry should not have been purged")
	}
}
