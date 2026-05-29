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
			// cc-connect priority: OpenId first, then UserId, then UnionId
			msg.UserID = strVal(event.Event.Sender.SenderId.OpenId)
			if msg.UserID == "" {
				msg.UserID = strVal(event.Event.Sender.SenderId.UserId)
			}
			if msg.UserID == "" {
				msg.UserID = strVal(event.Event.Sender.SenderId.UnionId)
			}
		}

		if event.Event.Message != nil {
			m := event.Event.Message
			msg.ID = strVal(m.MessageId)
			msg.ChatID = strVal(m.ChatId)
			msg.ThreadID = strVal(m.ThreadId)

			chatType := strVal(m.ChatType)
			msg.IsDirect = chatType == "p2p"

			// Detect message type
			msgType := strVal(m.MessageType)
			msg.HasImage = msgType == "image"
			msg.HasFile = msgType == "file"

			rawText := extractText(m)
			msg.Text = a.removeBotMention(rawText)
			msg.Mentioned = a.isBotMentioned(m)
		}
	}

	msg.ReplyTarget = core.ReplyTarget{
		ChatID:   msg.ChatID,
		ThreadID: msg.ThreadID,
	}

	if msg.ThreadID == "" {
		msg.ThreadID = msg.ID
		msg.ReplyTarget.ThreadID = msg.ID
	}

	return msg
}

// removeBotMention strips <at>user_id</at> tags for the bot itself.
func (a *Adapter) removeBotMention(text string) string {
	if a.botOpenID == "" {
		return strings.TrimSpace(text)
	}
	// Lark @mention format: <at user_id="ou_xxx">@BotName</at>
	text = strings.ReplaceAll(text, `<at user_id="`+a.botOpenID+`">@`+a.botName+`</at>`, "")
	text = strings.ReplaceAll(text, `<at user_id="`+a.botOpenID+`"></at>`, "")
	text = strings.ReplaceAll(text, "@"+a.botName, "")
	return strings.TrimSpace(text)
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

// isBotMentioned checks if the current bot is specifically mentioned.
// If botOpenID is not set, falls back to checking for any mention.
func (a *Adapter) isBotMentioned(m *larkim.EventMessage) bool {
	if len(m.Mentions) == 0 {
		return false
	}

	// If we don't know the bot's OpenID yet, fall back to any mention
	if a.botOpenID == "" {
		for _, mention := range m.Mentions {
			if mention != nil && mention.Name != nil && *mention.Name != "" {
				return true
			}
		}
		return false
	}

	// Check if any mention matches the bot's OpenID
	for _, mention := range m.Mentions {
		if mention == nil || mention.Id == nil {
			continue
		}
		mentionOpenID := strVal(mention.Id.OpenId)
		if mentionOpenID == a.botOpenID {
			return true
		}
	}
	return false
}

func strPtr(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
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
