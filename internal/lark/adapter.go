package lark

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/StefanIE1205/lark-agent-bridge/internal/core"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

const (
	platformName   = "lark"
	dedupTTL       = 5 * time.Minute
	dedupCleanup   = 1 * time.Minute
)

type dedupEntry struct {
	seen time.Time
}

type Adapter struct {
	appID     string
	appSecret string
	logger    *log.Logger
	wsClient  *larkws.Client
	apiClient *lark.Client
	dedup     map[string]dedupEntry
	dedupMu   sync.Mutex
}

func NewAdapter(appID, appSecret string, logger *log.Logger) *Adapter {
	if logger == nil {
		logger = log.New(os.Stderr, "[lark] ", log.LstdFlags)
	}
	return &Adapter{
		appID:     appID,
		appSecret: appSecret,
		logger:    logger,
		dedup:     make(map[string]dedupEntry),
	}
}

func (a *Adapter) Name() string {
	return platformName
}

func (a *Adapter) Start(ctx context.Context, handler core.MessageHandler) error {
	a.apiClient = lark.NewClient(a.appID, a.appSecret)

	dispatch := dispatcher.NewEventDispatcher("", "")
	dispatch.OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
		a.handleMessage(ctx, event, handler)
		return nil
	})

	a.wsClient = larkws.NewClient(a.appID, a.appSecret,
		larkws.WithEventHandler(dispatch),
		larkws.WithAutoReconnect(true),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
	)

	go a.dedupGC(ctx)

	a.logger.Printf("lark adapter starting (app_id=%s...)", a.appID[:min(8, len(a.appID))])

	if err := a.wsClient.Start(ctx); err != nil {
		return fmt.Errorf("lark: start client: %w", err)
	}

	return nil
}

func (a *Adapter) Stop(ctx context.Context) error {
	if a.wsClient != nil {
		a.wsClient.Close()
		a.logger.Printf("lark adapter stopped")
	}
	return nil
}

func (a *Adapter) Reply(ctx context.Context, target core.ReplyTarget, text string) error {
	return a.sendMessage(ctx, target, text)
}

func (a *Adapter) Update(ctx context.Context, target core.ReplyTarget, messageID string, text string) error {
	return fmt.Errorf("lark: Update not implemented")
}

func (a *Adapter) handleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1, handler core.MessageHandler) {
	msg := a.parseMessage(event)

	if msg.ID != "" && a.isDuplicate(msg.ID) {
		a.logger.Printf("skipping duplicate message: id=%s", msg.ID)
		return
	}

	a.logger.Printf("received message: id=%s chat=%s user=%s text=%q",
		msg.ID, msg.ChatID, msg.UserID, truncate(msg.Text, 80))

	if handler != nil {
		handler(ctx, msg)
	}
}

func (a *Adapter) isDuplicate(id string) bool {
	a.dedupMu.Lock()
	defer a.dedupMu.Unlock()

	if _, exists := a.dedup[id]; exists {
		return true
	}
	a.dedup[id] = dedupEntry{seen: time.Now()}
	return false
}

func (a *Adapter) dedupGC(ctx context.Context) {
	ticker := time.NewTicker(dedupCleanup)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.purgeOldEntries()
		}
	}
}

func (a *Adapter) purgeOldEntries() {
	a.dedupMu.Lock()
	defer a.dedupMu.Unlock()

	cutoff := time.Now().Add(-dedupTTL)
	for id, entry := range a.dedup {
		if entry.seen.Before(cutoff) {
			delete(a.dedup, id)
		}
	}
}
