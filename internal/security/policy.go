package security

type Policy struct {
	adminUserIDs   map[string]bool
	allowedChatIDs map[string]bool
	requireMention bool
	devAllowAll    bool
}

func NewPolicy(adminUserIDs []string, allowedChatIDs []string, requireMention bool, devAllowAll ...bool) *Policy {
	allowAll := false
	if len(devAllowAll) > 0 {
		allowAll = devAllowAll[0]
	}

	p := &Policy{
		adminUserIDs:   make(map[string]bool),
		allowedChatIDs: make(map[string]bool),
		requireMention: requireMention,
		devAllowAll:    allowAll,
	}
	for _, id := range adminUserIDs {
		if id != "" {
			p.adminUserIDs[id] = true
		}
	}
	for _, id := range allowedChatIDs {
		if id != "" {
			p.allowedChatIDs[id] = true
		}
	}
	return p
}

func (p *Policy) IsAdmin(userID string) bool {
	// Empty admin list: only allow all when devAllowAll is explicitly set
	if len(p.adminUserIDs) == 0 {
		return p.devAllowAll
	}
	return p.adminUserIDs[userID]
}

func (p *Policy) CanAccess(userID string, isDirect bool, chatID string, mentioned bool) bool {
	if userID == "" {
		return false
	}

	if isDirect {
		return p.IsAdmin(userID)
	}

	// Empty chat allowlist = allow all chats
	if len(p.allowedChatIDs) > 0 && !p.allowedChatIDs[chatID] {
		return false
	}

	if p.requireMention && !mentioned {
		return false
	}

	return true
}

func (p *Policy) CanExecute(userID string, cmdName string) bool {
	if p.IsAdmin(userID) {
		return true
	}

	switch cmdName {
	case "ping", "help", "status", "sessions", "log", "project", "agent":
		return true
	default:
		return false
	}
}

func PrivilegedCommands() []string {
	return []string{"bind", "ask", "stop", "approve", "deny"}
}
