package security

import (
	"testing"
)

func newTestPolicy() *Policy {
	return NewPolicy(
		[]string{"ou_admin"},
		[]string{"oc_allowed_group"},
		true,
	)
}

func TestIsAdmin(t *testing.T) {
	p := newTestPolicy()

	if !p.IsAdmin("ou_admin") {
		t.Error("ou_admin should be admin")
	}
	if p.IsAdmin("ou_rando") {
		t.Error("ou_rando should not be admin")
	}
}

func TestCanAccessPrivateChatAdmin(t *testing.T) {
	p := newTestPolicy()
	if !p.CanAccess("ou_admin", true, "", false) {
		t.Error("admin should be able to access private chat")
	}
}

func TestCanAccessPrivateChatNonAdmin(t *testing.T) {
	p := newTestPolicy()
	if p.CanAccess("ou_rando", true, "", false) {
		t.Error("non-admin should not be able to access private chat")
	}
}

func TestCanAccessPrivateChatEmptyUser(t *testing.T) {
	p := newTestPolicy()
	if p.CanAccess("", true, "", false) {
		t.Error("empty user should not be able to access")
	}
}

func TestCanAccessGroupChatAllowedMentioned(t *testing.T) {
	p := newTestPolicy()
	if !p.CanAccess("ou_rando", false, "oc_allowed_group", true) {
		t.Error("mentioned user in allowed group should have access")
	}
}

func TestCanAccessGroupChatAllowedNotMentioned(t *testing.T) {
	p := newTestPolicy()
	if p.CanAccess("ou_rando", false, "oc_allowed_group", false) {
		t.Error("non-mentioned user should not have access when require_mention=true")
	}
}

func TestCanAccessGroupChatAllowedNotMentionedDisabled(t *testing.T) {
	p := NewPolicy(
		[]string{"ou_admin"},
		[]string{"oc_allowed_group"},
		false,
	)
	if !p.CanAccess("ou_rando", false, "oc_allowed_group", false) {
		t.Error("non-mentioned user should have access when require_mention=false")
	}
}

func TestCanAccessGroupChatNotAllowed(t *testing.T) {
	p := newTestPolicy()
	if p.CanAccess("ou_admin", false, "oc_other_group", true) {
		t.Error("message from non-whitelisted group should be denied")
	}
}

func TestCanExecuteAdmin(t *testing.T) {
	p := newTestPolicy()

	privileged := []string{"bind", "ask", "stop", "approve", "deny"}
	readOnly := []string{"ping", "help", "status", "sessions", "log", "project", "agent"}

	for _, name := range privileged {
		if !p.CanExecute("ou_admin", name) {
			t.Errorf("admin should be able to run /%s", name)
		}
	}
	for _, name := range readOnly {
		if !p.CanExecute("ou_admin", name) {
			t.Errorf("admin should be able to run /%s", name)
		}
	}
}

func TestCanExecuteNonAdmin(t *testing.T) {
	p := newTestPolicy()

	readOnly := []string{"ping", "help", "status", "sessions", "log", "project", "agent"}
	privileged := []string{"bind", "ask", "stop", "approve", "deny"}

	for _, name := range readOnly {
		if !p.CanExecute("ou_rando", name) {
			t.Errorf("non-admin should be able to run /%s", name)
		}
	}
	for _, name := range privileged {
		if p.CanExecute("ou_rando", name) {
			t.Errorf("non-admin should NOT be able to run /%s", name)
		}
	}
}

func TestCanExecuteUnknownCommand(t *testing.T) {
	p := newTestPolicy()
	if p.CanExecute("ou_rando", "unknown") {
		t.Error("non-admin should not be able to run unknown commands")
	}
}

func TestCanExecuteEmptyCommand(t *testing.T) {
	p := newTestPolicy()
	if p.CanExecute("ou_rando", "") {
		t.Error("non-admin should not be able to run empty command")
	}
}

func TestPrivilegedCommands(t *testing.T) {
	cmds := PrivilegedCommands()
	if len(cmds) != 5 {
		t.Errorf("expected 5 privileged commands, got %d", len(cmds))
	}
	expected := map[string]bool{
		"bind": true, "ask": true, "stop": true, "approve": true, "deny": true,
	}
	for _, c := range cmds {
		if !expected[c] {
			t.Errorf("unexpected privileged command: %s", c)
		}
	}
}

func TestNewPolicyEmptyAdmins(t *testing.T) {
	p := NewPolicy(nil, nil, true)
	// Empty admin list = allow all users (cc-connect style)
	if !p.IsAdmin("ou_anyone") {
		t.Error("empty admin list should allow all users")
	}
}

func TestNewPolicyEmptyChats(t *testing.T) {
	p := NewPolicy([]string{"ou_admin"}, nil, true)
	// Empty chat allowlist = allow all group chats
	if !p.CanAccess("ou_admin", false, "oc_some_group", true) {
		t.Error("empty allowed_chat_ids should allow all group chats")
	}
}

func TestNewPolicySkipsEmptyStrings(t *testing.T) {
	p := NewPolicy(
		[]string{"ou_admin", "", "ou_also_admin"},
		[]string{"oc_chat", ""},
		true,
	)
	if !p.IsAdmin("ou_admin") {
		t.Error("ou_admin should be admin")
	}
	if !p.IsAdmin("ou_also_admin") {
		t.Error("ou_also_admin should be admin")
	}
	if p.IsAdmin("") {
		t.Error("empty string should not be admin")
	}
}
