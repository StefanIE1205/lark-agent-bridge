package lark

import (
	"encoding/json"
	"testing"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func makeEventMessage(content string) *larkim.EventMessage {
	return larkim.NewEventMessageBuilder().
		MessageId("msg_001").
		ChatId("oc_chat").
		ThreadId("ot_thread").
		ChatType("p2p").
		MessageType("text").
		Content(content).
		Build()
}

func TestExtractText(t *testing.T) {
	content := `{"text":"hello world"}`
	m := makeEventMessage(content)
	got := extractText(m)
	if got != "hello world" {
		t.Errorf("extractText = %q, want %q", got, "hello world")
	}
}

func TestExtractTextEmpty(t *testing.T) {
	m := makeEventMessage("")
	got := extractText(m)
	if got != "" {
		t.Errorf("extractText = %q, want empty", got)
	}
}

func TestExtractTextPlainJSON(t *testing.T) {
	// Not valid content JSON, but should not crash
	content := `just plain text, not json`
	m := makeEventMessage(content)
	got := extractText(m)
	if got != content {
		t.Errorf("extractText = %q, want %q", got, content)
	}
}

func TestStrVal(t *testing.T) {
	s := "hello"
	if strVal(&s) != "hello" {
		t.Error("strVal failed for non-nil")
	}
	if strVal(nil) != "" {
		t.Error("strVal failed for nil")
	}
}

func TestTruncate(t *testing.T) {
	got := truncate("hello world", 5)
	if got != "hello..." {
		t.Errorf("truncate = %q, want %q", got, "hello...")
	}
	got = truncate("hi", 10)
	if got != "hi" {
		t.Errorf("truncate = %q, want %q", got, "hi")
	}
}

func TestMin(t *testing.T) {
	if min(3, 5) != 3 {
		t.Error("min(3,5) != 3")
	}
	if min(5, 3) != 3 {
		t.Error("min(5,3) != 3")
	}
}

func TestParseMessageBasic(t *testing.T) {
	content, _ := json.Marshal(map[string]string{
		"text": "hello lark",
	})

	m := larkim.NewEventMessageBuilder().
		MessageId("msg_001").
		ChatId("oc_chat_123").
		ThreadId("ot_thread_456").
		ChatType("p2p").
		MessageType("text").
		Content(string(content)).
		Build()

	senderId := larkim.NewUserIdBuilder().UserId("ou_user_789").Build()

	sender := larkim.NewEventSenderBuilder().
		SenderId(senderId).
		Build()

	data := &larkim.P2MessageReceiveV1Data{
		Sender:  sender,
		Message: m,
	}

	event := &larkim.P2MessageReceiveV1{
		Event: data,
	}

	a := &Adapter{}
	msg := a.parseMessage(event)

	if msg.ID != "msg_001" {
		t.Errorf("ID = %q, want msg_001", msg.ID)
	}
	if msg.ChatID != "oc_chat_123" {
		t.Errorf("ChatID = %q, want oc_chat_123", msg.ChatID)
	}
	if msg.ThreadID != "ot_thread_456" {
		t.Errorf("ThreadID = %q, want ot_thread_456", msg.ThreadID)
	}
	if msg.UserID != "ou_user_789" {
		t.Errorf("UserID = %q, want ou_user_789", msg.UserID)
	}
	if msg.Text != "hello lark" {
		t.Errorf("Text = %q, want hello lark", msg.Text)
	}
	if !msg.IsDirect {
		t.Error("expected IsDirect = true for p2p chat")
	}
	if msg.Platform != "lark" {
		t.Errorf("Platform = %q, want lark", msg.Platform)
	}
}

func TestIsBotMentioned(t *testing.T) {
	botOpenID := "ou_bot_123"

	tests := []struct {
		name     string
		botOID   string
		mentions []*larkim.MentionEvent
		want     bool
	}{
		{
			name:   "bot is mentioned",
			botOID: botOpenID,
			mentions: []*larkim.MentionEvent{
				larkim.NewMentionEventBuilder().
					Name("Bot").
					Id(larkim.NewUserIdBuilder().OpenId(botOpenID).Build()).
					Build(),
			},
			want: true,
		},
		{
			name:   "other user mentioned, not bot",
			botOID: botOpenID,
			mentions: []*larkim.MentionEvent{
				larkim.NewMentionEventBuilder().
					Name("Alice").
					Id(larkim.NewUserIdBuilder().OpenId("ou_alice").Build()).
					Build(),
			},
			want: false,
		},
		{
			name:     "no mentions",
			botOID:   botOpenID,
			mentions: nil,
			want:     false,
		},
		{
			name:   "bot OpenID not set, fallback to any mention",
			botOID: "",
			mentions: []*larkim.MentionEvent{
				larkim.NewMentionEventBuilder().
					Name("Someone").
					Build(),
			},
			want: true,
		},
		{
			name:   "bot OpenID not set, no mentions",
			botOID: "",
			mentions: nil,
			want:     false,
		},
		{
			name:   "multiple mentions including bot",
			botOID: botOpenID,
			mentions: []*larkim.MentionEvent{
				larkim.NewMentionEventBuilder().
					Name("Alice").
					Id(larkim.NewUserIdBuilder().OpenId("ou_alice").Build()).
					Build(),
				larkim.NewMentionEventBuilder().
					Name("Bot").
					Id(larkim.NewUserIdBuilder().OpenId(botOpenID).Build()).
					Build(),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Adapter{botOpenID: tt.botOID}
			m := larkim.NewEventMessageBuilder().
				MessageId("msg_test").
				ChatId("oc_chat").
				ChatType("group").
				MessageType("text").
				Content(`{"text":"test"}`).
				Mentions(tt.mentions).
				Build()

			got := a.isBotMentioned(m)
			if got != tt.want {
				t.Errorf("isBotMentioned() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMessageThreadIDFallback(t *testing.T) {
	m := larkim.NewEventMessageBuilder().
		MessageId("msg_002").
		ChatId("oc_chat").
		ChatType("group").
		MessageType("text").
		Content(`{"text":"test"}`).
		Build()

	data := &larkim.P2MessageReceiveV1Data{
		Message: m,
	}

	event := &larkim.P2MessageReceiveV1{
		Event: data,
	}

	a := &Adapter{}
	msg := a.parseMessage(event)

	// ThreadID fallback to MessageID when ThreadID is empty
	if msg.ThreadID != "msg_002" {
		t.Errorf("ThreadID fallback = %q, want msg_002", msg.ThreadID)
	}
}
