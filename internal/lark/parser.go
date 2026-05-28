package lark

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/StefanIE1205/lark-agent-bridge/internal/core"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func (a *Adapter) parseMessage(event *larkim.P2MessageReceiveV1) core.Message {
	msg := core.Message{
		Platform:   platformName,
		ReceivedAt: time.Now(),
	}

	if event.Event != nil {
		if event.Event.Sender != nil && event.Event.Sender.SenderId != nil {
			msg.UserID = strVal(event.Event.Sender.SenderId.UserId)
			if msg.UserID == "" {
				msg.UserID = strVal(event.Event.Sender.SenderId.OpenId)
			}
		}

		if event.Event.Message != nil {
			m := event.Event.Message
			msg.ID = strVal(m.MessageId)
			msg.ChatID = strVal(m.ChatId)
			msg.ThreadID = strVal(m.ThreadId)

			chatType := strVal(m.ChatType)
			msg.IsDirect = chatType == "p2p"

			msg.Text = extractText(m)
			msg.Mentioned = isMentioned(m)
		}
	}

	// Build reply target from parsed data
	msg.ReplyTarget = core.ReplyTarget{
		ChatID:   msg.ChatID,
		ThreadID: msg.ThreadID,
	}

	// ThreadID fallback per spec
	if msg.ThreadID == "" {
		msg.ThreadID = msg.ID
		msg.ReplyTarget.ThreadID = msg.ID
	}

	return msg
}

func extractText(m *larkim.EventMessage) string {
	content := strVal(m.Content)
	if content == "" {
		return ""
	}

	var parsed struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return content
	}
	return parsed.Text
}

func isMentioned(m *larkim.EventMessage) bool {
	for _, mention := range m.Mentions {
		if mention != nil && mention.Name != nil && *mention.Name != "" {
			return true
		}
	}
	return false
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "")
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
