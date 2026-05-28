package security

type Policy struct {
	adminUserIDs   map[string]bool
	allowedChatIDs map[string]bool
	requireMention bool
}

func NewPolicy(adminUserIDs []string, allowedChatIDs []string, requireMention bool) *Policy {
	p := &Policy{
		adminUserIDs:   make(map[string]bool),
		allowedChatIDs: make(map[string]bool),
		requireMention: requireMention,
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
	return p.adminUserIDs[userID]
}

// CanAccess checks whether a message should be processed.
func (p *Policy) CanAccess(userID string, isDirect bool, chatID string, mentioned bool) bool {
	if userID == "" {
		return false
	}

	if isDirect {
		return p.IsAdmin(userID)
	}

	if !p.allowedChatIDs[chatID] {
		return false
	}

	if p.requireMention && !mentioned {
		return false
	}

	return true
}

// CanExecute checks whether a user can run a specific command.
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
