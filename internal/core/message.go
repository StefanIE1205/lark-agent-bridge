package core

import (
	"context"
	"time"
)

type Platform interface {
	Name() string
	Start(ctx context.Context, handler MessageHandler) error
	Reply(ctx context.Context, target ReplyTarget, text string) (string, error)
	Update(ctx context.Context, target ReplyTarget, messageID string, text string) error
	Stop(ctx context.Context) error
}

type MessageHandler func(ctx context.Context, msg Message)

type Message struct {
	ID          string
	Platform    string
	ChatID      string
	ThreadID    string
	UserID      string
	UserName    string
	Text        string
	Mentioned   bool
	IsDirect    bool
	HasImage    bool
	HasFile     bool
	ReplyTarget ReplyTarget
	ReceivedAt  time.Time
}

type ReplyTarget struct {
	ChatID    string
	ThreadID  string
	MessageID string
}
